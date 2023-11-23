// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// MConf is an application configuration system.
// It believes that applications can be configured a variety of ways
// via flags, ENVIRONMENT variables, configuration files retrieved
// from the file system, or a  key/value store.

// Each item takes precedence over the item below it:

// overrides
// flag
// env
// config
// key/value store
// default

package mconf

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"mlib.com/mrun/cast"

	"mlib.com/mrun/mconf/internal/encoding"
	"mlib.com/mrun/mconf/internal/encoding/ini"
	"mlib.com/mrun/mconf/internal/encoding/json"
	"mlib.com/mrun/mconf/internal/encoding/yaml"
)

// ConfigMarshalError happens when failing to marshal the configuration.
type ConfigMarshalError struct {
	err error
}

// Error returns the formatted configuration error.
func (e ConfigMarshalError) Error() string {
	return fmt.Sprintf("While marshaling config: %s", e.err.Error())
}

var v *MConf

func init() {
	v = New()
}

// UnsupportedConfigError denotes encountering an unsupported
// configuration filetype.
type UnsupportedConfigError string

// Error returns the formatted configuration error.
func (str UnsupportedConfigError) Error() string {
	return fmt.Sprintf("Unsupported Config Type %q", string(str))
}

// ConfigFileNotFoundError denotes failing to find configuration file.
type ConfigFileNotFoundError struct {
	name, locations string
}

// Error returns the formatted configuration error.
func (fnfe ConfigFileNotFoundError) Error() string {
	return fmt.Sprintf("Config File %q Not Found in %q", fnfe.name, fnfe.locations)
}

// ConfigFileAlreadyExistsError denotes failure to write new configuration file.
type ConfigFileAlreadyExistsError string

// Error returns the formatted error when configuration already exists.
func (faee ConfigFileAlreadyExistsError) Error() string {
	return fmt.Sprintf("Config File %q Already Exists", string(faee))
}

// MConf is a prioritized configuration registry. It
// maintains a set of configuration sources, fetches
// values to populate those, and provides them according
// to the source's priority.
// The priority of the sources is the following:
// 1. overrides
// 2. flags
// 3. env. variables
// 4. config file
// 5. key/value store
// 6. defaults
//
// For example, if values from the following sources were loaded:
//
//	Defaults : {
//		"secret": "",
//		"user": "default",
//		"endpoint": "https://localhost"
//	}
//	Config : {
//		"user": "root"
//		"secret": "defaultsecret"
//	}
//	Env : {
//		"secret": "somesecretkey"
//	}
//
// The resulting config will have the following values:
//
//	{
//		"secret": "somesecretkey",
//		"user": "root",
//		"endpoint": "https://localhost"
//	}
//
// Note: Vipers are not safe for concurrent Get() and Set() operations.
type MConf struct {
	// Delimiter that separates a list of keys
	// used to access a nested value in one go
	keyDelim string

	// A set of paths to look for the config file in
	configPaths []string

	// Name of file to look for inside the path
	configName        string
	configFile        string
	configType        string
	configPermissions os.FileMode

	// Specific commands for ini parsing
	iniLoadOptions ini.LoadOptions

	config         map[string]interface{}
	override       map[string]interface{}
	defaults       map[string]interface{}
	kvstore        map[string]interface{}
	aliases        map[string]string
	typeByDefValue bool

	// TODO: should probably be protected with a mutex
	encoderRegistry *encoding.EncoderRegistry
	decoderRegistry *encoding.DecoderRegistry
}

// New returns an initialized MConf instance.
func New() *MConf {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	v := new(MConf)
	v.keyDelim = "."
	v.configName = "config"
	v.configPermissions = os.FileMode(0o644)
	v.config = make(map[string]interface{})
	v.override = make(map[string]interface{})
	v.defaults = make(map[string]interface{})
	v.kvstore = make(map[string]interface{})
	v.aliases = make(map[string]string)
	v.typeByDefValue = false

	v.resetEncoding()

	return v
}

// Option configures MConf using the functional options paradigm popularized by Rob Pike and Dave Cheney.
// If you're unfamiliar with this style,
// see https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
type Option interface {
	apply(v *MConf)
}

type optionFunc func(v *MConf)

func (fn optionFunc) apply(v *MConf) {
	fn(v)
}

// KeyDelimiter sets the delimiter used for determining key parts.
// By default it's value is ".".
func KeyDelimiter(d string) Option {
	return optionFunc(func(v *MConf) {
		v.keyDelim = d
	})
}

// StringReplacer applies a set of replacements to a string.
type StringReplacer interface {
	// Replace returns a copy of s with all replacements performed.
	Replace(s string) string
}

// NewWithOptions creates a new MConf instance.
func NewWithOptions(opts ...Option) *MConf {
	v := New()

	for _, opt := range opts {
		opt.apply(v)
	}

	v.resetEncoding()

	return v
}

// Reset is intended for testing, will reset all to default settings.
// In the public interface for the viper package so applications
// can use it in their testing as well.
func Reset() {
	v = New()
	SupportedExts = []string{"json", "yaml", "yml", "ini"}
}

// TODO: make this lazy initialization instead
func (v *MConf) resetEncoding() {
	encoderRegistry := encoding.NewEncoderRegistry()
	decoderRegistry := encoding.NewDecoderRegistry()

	{
		codec := yaml.Codec{}

		encoderRegistry.RegisterEncoder("yaml", codec)
		decoderRegistry.RegisterDecoder("yaml", codec)

		encoderRegistry.RegisterEncoder("yml", codec)
		decoderRegistry.RegisterDecoder("yml", codec)
	}

	{
		codec := json.Codec{}

		encoderRegistry.RegisterEncoder("json", codec)
		decoderRegistry.RegisterDecoder("json", codec)
	}

	{
		codec := ini.Codec{
			KeyDelimiter: v.keyDelim,
			LoadOptions:  v.iniLoadOptions,
		}

		encoderRegistry.RegisterEncoder("ini", codec)
		decoderRegistry.RegisterDecoder("ini", codec)
	}

	v.encoderRegistry = encoderRegistry
	v.decoderRegistry = decoderRegistry
}

// SupportedExts are universally supported extensions.
var SupportedExts = []string{"json", "yaml", "yml", "ini"}

// SetConfigFile explicitly defines the path, name and extension of the config file.
// MConf will use this and not check any of the config paths.
func SetConfigFile(in string) { v.SetConfigFile(in) }

func (v *MConf) SetConfigFile(in string) {
	if in != "" {
		v.configFile = in
	}
}

// ConfigFileUsed returns the file used to populate the config registry.
func ConfigFileUsed() string            { return v.ConfigFileUsed() }
func (v *MConf) ConfigFileUsed() string { return v.configFile }

// AddConfigPath adds a path for MConf to search for the config file in.
// Can be called multiple times to define multiple search paths.
func AddConfigPath(in string) { v.AddConfigPath(in) }

func (v *MConf) AddConfigPath(in string) {
	if in != "" {
		absin := absPathify(in)

		log.Println("[I]adding path to search paths", "path", absin)
		if !stringInSlice(absin, v.configPaths) {
			v.configPaths = append(v.configPaths, absin)
		}
	}
}

// searchMap recursively searches for a value for path in source map.
// Returns nil if not found.
// Note: This assumes that the path entries and map keys are lower cased.
func (v *MConf) searchMap(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	next, ok := source[path[0]]
	if ok {
		// Fast path
		if len(path) == 1 {
			return next
		}

		// Nested case
		switch next := next.(type) {
		case map[interface{}]interface{}:
			return v.searchMap(cast.ToStringMap(next), path[1:])
		case map[string]interface{}:
			// Type assertion is safe here since it is only reached
			// if the type of `next` is the same as the type being asserted
			return v.searchMap(next, path[1:])
		default:
			// got a value but nested key expected, return "nil" for not found
			return nil
		}
	}
	return nil
}

// searchIndexableWithPathPrefixes recursively searches for a value for path in source map/slice.
//
// While searchMap() considers each path element as a single map key or slice index, this
// function searches for, and prioritizes, merged path elements.
// e.g., if in the source, "foo" is defined with a sub-key "bar", and "foo.bar"
// is also defined, this latter value is returned for path ["foo", "bar"].
//
// This should be useful only at config level (other maps may not contain dots
// in their keys).
//
// Note: This assumes that the path entries and map keys are lower cased.
func (v *MConf) searchIndexableWithPathPrefixes(source interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	// search for path prefixes, starting from the longest one
	for i := len(path); i > 0; i-- {
		prefixKey := strings.ToLower(strings.Join(path[0:i], v.keyDelim))

		var val interface{}
		switch sourceIndexable := source.(type) {
		case []interface{}:
			val = v.searchSliceWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		case map[string]interface{}:
			val = v.searchMapWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		}
		if val != nil {
			return val
		}
	}

	// not found
	return nil
}

// searchSliceWithPathPrefixes searches for a value for path in sourceSlice
//
// This function is part of the searchIndexableWithPathPrefixes recurring search and
// should not be called directly from functions other than searchIndexableWithPathPrefixes.
func (v *MConf) searchSliceWithPathPrefixes(
	sourceSlice []interface{},
	prefixKey string,
	pathIndex int,
	path []string,
) interface{} {
	// if the prefixKey is not a number or it is out of bounds of the slice
	index, err := strconv.Atoi(prefixKey)
	if err != nil || len(sourceSlice) <= index {
		return nil
	}

	next := sourceSlice[index]

	// Fast path
	if pathIndex == len(path) {
		return next
	}

	switch n := next.(type) {
	case map[interface{}]interface{}:
		return v.searchIndexableWithPathPrefixes(cast.ToStringMap(n), path[pathIndex:])
	case map[string]interface{}, []interface{}:
		return v.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

// searchMapWithPathPrefixes searches for a value for path in sourceMap
//
// This function is part of the searchIndexableWithPathPrefixes recurring search and
// should not be called directly from functions other than searchIndexableWithPathPrefixes.
func (v *MConf) searchMapWithPathPrefixes(
	sourceMap map[string]interface{},
	prefixKey string,
	pathIndex int,
	path []string,
) interface{} {
	next, ok := sourceMap[prefixKey]
	if !ok {
		return nil
	}

	// Fast path
	if pathIndex == len(path) {
		return next
	}

	// Nested case
	switch n := next.(type) {
	case map[interface{}]interface{}:
		return v.searchIndexableWithPathPrefixes(cast.ToStringMap(n), path[pathIndex:])
	case map[string]interface{}, []interface{}:
		return v.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

// isPathShadowedInDeepMap makes sure the given path is not shadowed somewhere
// on its path in the map.
// e.g., if "foo.bar" has a value in the given map, it “shadows”
//
//	"foo.bar.baz" in a lower-priority map
func (v *MConf) isPathShadowedInDeepMap(path []string, m map[string]interface{}) string {
	var parentVal interface{}
	for i := 1; i < len(path); i++ {
		parentVal = v.searchMap(m, path[0:i])
		if parentVal == nil {
			// not found, no need to add more path elements
			return ""
		}
		switch parentVal.(type) {
		case map[interface{}]interface{}:
			continue
		case map[string]interface{}:
			continue
		default:
			// parentVal is a regular value which shadows "path"
			return strings.Join(path[0:i], v.keyDelim)
		}
	}
	return ""
}

// SetTypeByDefaultValue enables or disables the inference of a key value's
// type when the Get function is used based upon a key's default value as
// opposed to the value returned based on the normal fetch logic.
//
// For example, if a key has a default value of []string{} and the same key
// is set via an environment variable to "a b c", a call to the Get function
// would return a string slice for the key if the key's type is inferred by
// the default value and the Get function would return:
//
//	[]string {"a", "b", "c"}
//
// Otherwise the Get function would return:
//
//	"a b c"
func SetTypeByDefaultValue(enable bool) { v.SetTypeByDefaultValue(enable) }

func (v *MConf) SetTypeByDefaultValue(enable bool) {
	v.typeByDefValue = enable
}

// GetViper gets the global MConf instance.
func GetViper() *MConf {
	return v
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. MConf will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{} { return v.Get(key) }

func (v *MConf) Get(key string) interface{} {
	lcaseKey := strings.ToLower(key)
	val := v.find(lcaseKey, true)
	if val == nil {
		return nil
	}

	if v.typeByDefValue {
		// TODO(bep) this branch isn't covered by a single test.
		valType := val
		path := strings.Split(lcaseKey, v.keyDelim)
		defVal := v.searchMap(v.defaults, path)
		if defVal != nil {
			valType = defVal
		}

		switch valType.(type) {
		case bool:
			return cast.ToBool(val)
		case string:
			return cast.ToString(val)
		case int32, int16, int8, int:
			return cast.ToInt(val)
		case uint:
			return cast.ToUint(val)
		case uint32:
			return cast.ToUint32(val)
		case uint64:
			return cast.ToUint64(val)
		case int64:
			return cast.ToInt64(val)
		case float64, float32:
			return cast.ToFloat64(val)
		case time.Time:
			return cast.ToTime(val)
		case time.Duration:
			return cast.ToDuration(val)
		case []string:
			return cast.ToStringSlice(val)
		case []int:
			return cast.ToIntSlice(val)
		}
	}

	return val
}

// Sub returns new MConf instance representing a sub tree of this instance.
// Sub is case-insensitive for a key.
func Sub(key string) *MConf { return v.Sub(key) }

func (v *MConf) Sub(key string) *MConf {
	subv := New()
	data := v.Get(key)
	if data == nil {
		return nil
	}

	if reflect.TypeOf(data).Kind() == reflect.Map {
		subv.config = cast.ToStringMap(data)
		return subv
	}
	return nil
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string { return v.GetString(key) }

func (v *MConf) GetString(key string) string {
	return cast.ToString(v.Get(key))
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool { return v.GetBool(key) }

func (v *MConf) GetBool(key string) bool {
	return cast.ToBool(v.Get(key))
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int { return v.GetInt(key) }

func (v *MConf) GetInt(key string) int {
	return cast.ToInt(v.Get(key))
}

// GetInt32 returns the value associated with the key as an integer.
func GetInt32(key string) int32 { return v.GetInt32(key) }

func (v *MConf) GetInt32(key string) int32 {
	return cast.ToInt32(v.Get(key))
}

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(key string) int64 { return v.GetInt64(key) }

func (v *MConf) GetInt64(key string) int64 {
	return cast.ToInt64(v.Get(key))
}

// GetUint returns the value associated with the key as an unsigned integer.
func GetUint(key string) uint { return v.GetUint(key) }

func (v *MConf) GetUint(key string) uint {
	return cast.ToUint(v.Get(key))
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func GetUint32(key string) uint32 { return v.GetUint32(key) }

func (v *MConf) GetUint32(key string) uint32 {
	return cast.ToUint32(v.Get(key))
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(key string) uint64 { return v.GetUint64(key) }

func (v *MConf) GetUint64(key string) uint64 {
	return cast.ToUint64(v.Get(key))
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64 { return v.GetFloat64(key) }

func (v *MConf) GetFloat64(key string) float64 {
	return cast.ToFloat64(v.Get(key))
}

// GetTime returns the value associated with the key as time.
func GetTime(key string) time.Time { return v.GetTime(key) }

func (v *MConf) GetTime(key string) time.Time {
	return cast.ToTime(v.Get(key))
}

// GetDuration returns the value associated with the key as a duration.
func GetDuration(key string) time.Duration { return v.GetDuration(key) }

func (v *MConf) GetDuration(key string) time.Duration {
	return cast.ToDuration(v.Get(key))
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func GetIntSlice(key string) []int { return v.GetIntSlice(key) }

func (v *MConf) GetIntSlice(key string) []int {
	return cast.ToIntSlice(v.Get(key))
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(key string) []string { return v.GetStringSlice(key) }

func (v *MConf) GetStringSlice(key string) []string {
	return cast.ToStringSlice(v.Get(key))
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func GetStringMap(key string) map[string]interface{} { return v.GetStringMap(key) }

func (v *MConf) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(v.Get(key))
}

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapString(key string) map[string]string { return v.GetStringMapString(key) }

func (v *MConf) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(v.Get(key))
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func GetStringMapStringSlice(key string) map[string][]string { return v.GetStringMapStringSlice(key) }

func (v *MConf) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(v.Get(key))
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
func GetSizeInBytes(key string) uint { return v.GetSizeInBytes(key) }

func (v *MConf) GetSizeInBytes(key string) uint {
	sizeStr := cast.ToString(v.Get(key))
	return parseSizeInBytes(sizeStr)
}

// Given a key, find the value.
//
// MConf will check to see if an alias exists first.
// MConf will then check in the following order:
// flag, env, config file, key/value store.
// Lastly, if no value was found and flagDefault is true, and if the key
// corresponds to a flag, the flag's default value is returned.
//
// Note: this assumes a lower-cased key given.
func (v *MConf) find(lcaseKey string, flagDefault bool) interface{} {
	var (
		val interface{}
		// exists bool
		path   = strings.Split(lcaseKey, v.keyDelim)
		nested = len(path) > 1
	)

	// compute the path through the nested maps to the nested value
	if nested && v.isPathShadowedInDeepMap(path, castMapStringToMapInterface(v.aliases)) != "" {
		return nil
	}

	// if the requested key is an alias, then return the proper key
	lcaseKey = v.realKey(lcaseKey)
	path = strings.Split(lcaseKey, v.keyDelim)
	nested = len(path) > 1

	// Set() override first
	val = v.searchMap(v.override, path)
	if val != nil {
		return val
	}
	if nested && v.isPathShadowedInDeepMap(path, v.override) != "" {
		return nil
	}

	// Config file next
	val = v.searchIndexableWithPathPrefixes(v.config, path)
	if val != nil {
		return val
	}
	if nested && v.isPathShadowedInDeepMap(path, v.config) != "" {
		return nil
	}

	// K/V store next
	val = v.searchMap(v.kvstore, path)
	if val != nil {
		return val
	}
	if nested && v.isPathShadowedInDeepMap(path, v.kvstore) != "" {
		return nil
	}

	// Default next
	val = v.searchMap(v.defaults, path)
	if val != nil {
		return val
	}
	if nested && v.isPathShadowedInDeepMap(path, v.defaults) != "" {
		return nil
	}

	return nil
}

func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

// mostly copied from pflag's implementation of this operation here https://github.com/spf13/pflag/blob/master/string_to_string.go#L79
// alterations are: errors are swallowed, map[string]interface{} is returned in order to enable cast.ToStringMap
func stringToStringConv(val string) interface{} {
	val = strings.Trim(val, "[]")
	// An empty string would cause an empty map
	if len(val) == 0 {
		return map[string]interface{}{}
	}
	r := csv.NewReader(strings.NewReader(val))
	ss, err := r.Read()
	if err != nil {
		return nil
	}
	out := make(map[string]interface{}, len(ss))
	for _, pair := range ss {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil
		}
		out[kv[0]] = kv[1]
	}
	return out
}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key.
func IsSet(key string) bool { return v.IsSet(key) }

func (v *MConf) IsSet(key string) bool {
	lcaseKey := strings.ToLower(key)
	val := v.find(lcaseKey, false)
	return val != nil
}

// RegisterAlias creates an alias that provides another accessor for the same key.
// This enables one to change a name without breaking the application.
func RegisterAlias(alias string, key string) { v.RegisterAlias(alias, key) }

func (v *MConf) RegisterAlias(alias string, key string) {
	v.registerAlias(alias, strings.ToLower(key))
}

func (v *MConf) registerAlias(alias string, key string) {
	alias = strings.ToLower(alias)
	if alias != key && alias != v.realKey(key) {
		_, exists := v.aliases[alias]

		if !exists {
			// if we alias something that exists in one of the maps to another
			// name, we'll never be able to get that value using the original
			// name, so move the config value to the new realkey.
			if val, ok := v.config[alias]; ok {
				delete(v.config, alias)
				v.config[key] = val
			}
			if val, ok := v.kvstore[alias]; ok {
				delete(v.kvstore, alias)
				v.kvstore[key] = val
			}
			if val, ok := v.defaults[alias]; ok {
				delete(v.defaults, alias)
				v.defaults[key] = val
			}
			if val, ok := v.override[alias]; ok {
				delete(v.override, alias)
				v.override[key] = val
			}
			v.aliases[alias] = key
		}
	} else {
		log.Println("[W]creating circular reference alias", "alias", alias, "key", key, "real_key", v.realKey(key))
	}
}

func (v *MConf) realKey(key string) string {
	newkey, exists := v.aliases[key]
	if exists {
		log.Println("[D]key is an alias", "alias", key, "to", newkey)

		return v.realKey(newkey)
	}
	return key
}

// InConfig checks to see if the given key (or an alias) is in the config file.
func InConfig(key string) bool { return v.InConfig(key) }

func (v *MConf) InConfig(key string) bool {
	lcaseKey := strings.ToLower(key)

	// if the requested key is an alias, then return the proper key
	lcaseKey = v.realKey(lcaseKey)
	path := strings.Split(lcaseKey, v.keyDelim)

	return v.searchIndexableWithPathPrefixes(v.config, path) != nil
}

// SetDefault sets the default value for this key.
// SetDefault is case-insensitive for a key.
// Default only used when no value is provided by the user via flag, config or ENV.
func SetDefault(key string, value interface{}) { v.SetDefault(key, value) }

func (v *MConf) SetDefault(key string, value interface{}) {
	// If alias passed in, then set the proper default
	key = v.realKey(strings.ToLower(key))
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, v.keyDelim)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(v.defaults, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// Set sets the value for the key in the override register.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func Set(key string, value interface{}) { v.Set(key, value) }

func (v *MConf) Set(key string, value interface{}) {
	// If alias passed in, then set the proper override
	key = v.realKey(strings.ToLower(key))
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, v.keyDelim)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(v.override, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// ReadInConfig will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
func ReadInConfig() error { return v.ReadInConfig() }

func (v *MConf) ReadInConfig() error {
	log.Println("[I]attempting to read in config file")
	filename, err := v.getConfigFile()
	if err != nil {
		return err
	}

	if !stringInSlice(v.getConfigType(), SupportedExts) {
		return UnsupportedConfigError(v.getConfigType())
	}

	log.Println("[D]reading file", "file", filename)
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	config := make(map[string]interface{})

	err = v.unmarshalReader(bytes.NewReader(file), config)
	if err != nil {
		return err
	}

	v.config = config
	return nil
}

// MergeInConfig merges a new configuration with an existing config.
func MergeInConfig() error { return v.MergeInConfig() }

func (v *MConf) MergeInConfig() error {
	log.Println("[I]attempting to merge in config file")
	filename, err := v.getConfigFile()
	if err != nil {
		return err
	}

	if !stringInSlice(v.getConfigType(), SupportedExts) {
		return UnsupportedConfigError(v.getConfigType())
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return v.MergeConfig(bytes.NewReader(file))
}

// ReadConfig will read a configuration file, setting existing keys to nil if the
// key does not exist in the file.
func ReadConfig(in io.Reader) error { return v.ReadConfig(in) }

func (v *MConf) ReadConfig(in io.Reader) error {
	v.config = make(map[string]interface{})
	return v.unmarshalReader(in, v.config)
}

// MergeConfig merges a new configuration with an existing config.
func MergeConfig(in io.Reader) error { return v.MergeConfig(in) }

func (v *MConf) MergeConfig(in io.Reader) error {
	cfg := make(map[string]interface{})
	if err := v.unmarshalReader(in, cfg); err != nil {
		return err
	}
	return v.MergeConfigMap(cfg)
}

// MergeConfigMap merges the configuration from the map given with an existing config.
// Note that the map given may be modified.
func MergeConfigMap(cfg map[string]interface{}) error { return v.MergeConfigMap(cfg) }

func (v *MConf) MergeConfigMap(cfg map[string]interface{}) error {
	if v.config == nil {
		v.config = make(map[string]interface{})
	}
	insensitiviseMap(cfg)
	mergeMaps(cfg, v.config, nil)
	return nil
}

// WriteConfig writes the current configuration to a file.
func WriteConfig() error { return v.WriteConfig() }

func (v *MConf) WriteConfig() error {
	filename, err := v.getConfigFile()
	if err != nil {
		return err
	}
	return v.writeConfig(filename, true)
}

// SafeWriteConfig writes current configuration to file only if the file does not exist.
func SafeWriteConfig() error { return v.SafeWriteConfig() }

func (v *MConf) SafeWriteConfig() error {
	if len(v.configPaths) < 1 {
		return errors.New("missing configuration for 'configPath'")
	}
	return v.SafeWriteConfigAs(filepath.Join(v.configPaths[0], v.configName+"."+v.configType))
}

// WriteConfigAs writes current configuration to a given filename.
func WriteConfigAs(filename string) error { return v.WriteConfigAs(filename) }

func (v *MConf) WriteConfigAs(filename string) error {
	return v.writeConfig(filename, true)
}

// SafeWriteConfigAs writes current configuration to a given filename if it does not exist.
func SafeWriteConfigAs(filename string) error { return v.SafeWriteConfigAs(filename) }

func (v *MConf) SafeWriteConfigAs(filename string) error {
	alreadyExists, err := OsExists(filename)
	if alreadyExists && err == nil {
		return ConfigFileAlreadyExistsError(filename)
	}
	return v.writeConfig(filename, false)
}

func (v *MConf) writeConfig(filename string, force bool) error {
	log.Println("[I]attempting to write configuration to file")

	var configType string

	ext := filepath.Ext(filename)
	if ext != "" && ext != filepath.Base(filename) {
		configType = ext[1:]
	} else {
		configType = v.configType
	}
	if configType == "" {
		return fmt.Errorf("config type could not be determined for %s", filename)
	}

	if !stringInSlice(configType, SupportedExts) {
		return UnsupportedConfigError(configType)
	}
	if v.config == nil {
		v.config = make(map[string]interface{})
	}
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	if !force {
		flags |= os.O_EXCL
	}

	f, err := os.OpenFile(filename, flags, v.configPermissions)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := v.marshalWriter(f, configType); err != nil {
		return err
	}

	return f.Sync()
}

// Unmarshal a Reader into a map.
// Should probably be an unexported function.
func unmarshalReader(in io.Reader, c map[string]interface{}) error {
	return v.unmarshalReader(in, c)
}

func (v *MConf) unmarshalReader(in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	switch format := strings.ToLower(v.getConfigType()); format {
	case "yaml", "yml", "json", "ini":
		err := v.decoderRegistry.Decode(format, buf.Bytes(), c)
		if err != nil {
			return ConfigParseError{err}
		}
	}

	insensitiviseMap(c)
	return nil
}

// Marshal a map into Writer.
func (v *MConf) marshalWriter(f *os.File, configType string) error {
	c := v.AllSettings()
	switch configType {
	case "yaml", "yml", "json", "ini":
		b, err := v.encoderRegistry.Encode(configType, c)
		if err != nil {
			return ConfigMarshalError{err}
		}

		_, err = f.WriteString(string(b))
		if err != nil {
			return ConfigMarshalError{err}
		}
	}
	return nil
}

func keyExists(k string, m map[string]interface{}) string {
	lk := strings.ToLower(k)
	for mk := range m {
		lmk := strings.ToLower(mk)
		if lmk == lk {
			return mk
		}
	}
	return ""
}

func castToMapStringInterface(
	src map[interface{}]interface{},
) map[string]interface{} {
	tgt := map[string]interface{}{}
	for k, v := range src {
		tgt[fmt.Sprintf("%v", k)] = v
	}
	return tgt
}

func castMapStringSliceToMapInterface(src map[string][]string) map[string]interface{} {
	tgt := map[string]interface{}{}
	for k, v := range src {
		tgt[k] = v
	}
	return tgt
}

func castMapStringToMapInterface(src map[string]string) map[string]interface{} {
	tgt := map[string]interface{}{}
	for k, v := range src {
		tgt[k] = v
	}
	return tgt
}

// mergeMaps merges two maps. The `itgt` parameter is for handling go-yaml's
// insistence on parsing nested structures as `map[interface{}]interface{}`
// instead of using a `string` as the key for nest structures beyond one level
// deep. Both map types are supported as there is a go-yaml fork that uses
// `map[string]interface{}` instead.
func mergeMaps(
	src, tgt map[string]interface{}, itgt map[interface{}]interface{},
) {
	for sk, sv := range src {
		tk := keyExists(sk, tgt)
		if tk == "" {
			log.Println("[D]", "tk", "\"\"", fmt.Sprintf("tgt[%s]", sk), sv)
			tgt[sk] = sv
			if itgt != nil {
				itgt[sk] = sv
			}
			continue
		}

		tv, ok := tgt[tk]
		if !ok {
			log.Println("[D]", fmt.Sprintf("ok[%s]", tk), false, fmt.Sprintf("tgt[%s]", sk), sv)
			tgt[sk] = sv
			if itgt != nil {
				itgt[sk] = sv
			}
			continue
		}

		svType := reflect.TypeOf(sv)
		tvType := reflect.TypeOf(tv)

		log.Println("[D]processing",
			"key", sk,
			"st", svType,
			"tt", tvType,
			"sv", sv,
			"tv", tv,
		)

		switch ttv := tv.(type) {
		case map[interface{}]interface{}:
			log.Println("[D]merging maps (must convert)")
			tsv, ok := sv.(map[interface{}]interface{})
			if !ok {
				log.Println("[E]Could not cast sv to map[interface{}]interface{}",
					"key", sk,
					"st", svType,
					"tt", tvType,
					"sv", sv,
					"tv", tv,
				)
				continue
			}

			ssv := castToMapStringInterface(tsv)
			stv := castToMapStringInterface(ttv)
			mergeMaps(ssv, stv, ttv)
		case map[string]interface{}:
			log.Println("[D]merging maps")
			tsv, ok := sv.(map[string]interface{})
			if !ok {
				log.Println("[E]Could not cast sv to map[string]interface{}",
					"key", sk,
					"st", svType,
					"tt", tvType,
					"sv", sv,
					"tv", tv,
				)
				continue
			}
			mergeMaps(tsv, ttv, nil)
		default:
			log.Println("[D]setting value")
			tgt[tk] = sv
			if itgt != nil {
				itgt[tk] = sv
			}
		}
	}
}

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are returned with a v.keyDelim separator
func AllKeys() []string { return v.AllKeys() }

func (v *MConf) AllKeys() []string {
	m := map[string]bool{}
	// add all paths, by order of descending priority to ensure correct shadowing
	m = v.flattenAndMergeMap(m, castMapStringToMapInterface(v.aliases), "")
	m = v.flattenAndMergeMap(m, v.override, "")
	m = v.flattenAndMergeMap(m, v.config, "")
	m = v.flattenAndMergeMap(m, v.kvstore, "")
	m = v.flattenAndMergeMap(m, v.defaults, "")

	// convert set of paths to list
	a := make([]string, 0, len(m))
	for x := range m {
		a = append(a, x)
	}
	return a
}

// flattenAndMergeMap recursively flattens the given map into a map[string]bool
// of key paths (used as a set, easier to manipulate than a []string):
//   - each path is merged into a single key string, delimited with v.keyDelim
//   - if a path is shadowed by an earlier value in the initial shadow map,
//     it is skipped.
//
// The resulting set of paths is merged to the given shadow set at the same time.
func (v *MConf) flattenAndMergeMap(shadow map[string]bool, m map[string]interface{}, prefix string) map[string]bool {
	if shadow != nil && prefix != "" && shadow[prefix] {
		// prefix is shadowed => nothing more to flatten
		return shadow
	}
	if shadow == nil {
		shadow = make(map[string]bool)
	}

	var m2 map[string]interface{}
	if prefix != "" {
		prefix += v.keyDelim
	}
	for k, val := range m {
		fullKey := prefix + k
		switch val.(type) {
		case map[string]interface{}:
			m2 = val.(map[string]interface{})
		case map[interface{}]interface{}:
			m2 = cast.ToStringMap(val)
		default:
			// immediate value
			shadow[strings.ToLower(fullKey)] = true
			continue
		}
		// recursively merge to shadow map
		shadow = v.flattenAndMergeMap(shadow, m2, fullKey)
	}
	return shadow
}

// mergeFlatMap merges the given maps, excluding values of the second map
// shadowed by values from the first map.
func (v *MConf) mergeFlatMap(shadow map[string]bool, m map[string]interface{}) map[string]bool {
	// scan keys
outer:
	for k := range m {
		path := strings.Split(k, v.keyDelim)
		// scan intermediate paths
		var parentKey string
		for i := 1; i < len(path); i++ {
			parentKey = strings.Join(path[0:i], v.keyDelim)
			if shadow[parentKey] {
				// path is shadowed, continue
				continue outer
			}
		}
		// add key
		shadow[strings.ToLower(k)] = true
	}
	return shadow
}

// AllSettings merges all settings and returns them as a map[string]interface{}.
func AllSettings() map[string]interface{} { return v.AllSettings() }

func (v *MConf) AllSettings() map[string]interface{} {
	m := map[string]interface{}{}
	// start from the list of keys, and construct the map one value at a time
	for _, k := range v.AllKeys() {
		value := v.Get(k)
		if value == nil {
			// should not happen, since AllKeys() returns only keys holding a value,
			// check just in case anything changes
			continue
		}
		path := strings.Split(k, v.keyDelim)
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := deepSearch(m, path[0:len(path)-1])
		// set innermost value
		deepestMap[lastKey] = value
	}
	return m
}

// SetConfigName sets name for the config file.
// Does not include extension.
func SetConfigName(in string) { v.SetConfigName(in) }

func (v *MConf) SetConfigName(in string) {
	if in != "" {
		v.configName = in
		v.configFile = ""
	}
}

// SetConfigType sets the type of the configuration returned by the
// source, e.g. "json".
func SetConfigType(in string) { v.SetConfigType(in) }

func (v *MConf) SetConfigType(in string) {
	if in != "" {
		v.configType = in
	}
}

// SetConfigPermissions sets the permissions for the config file.
func SetConfigPermissions(perm os.FileMode) { v.SetConfigPermissions(perm) }

func (v *MConf) SetConfigPermissions(perm os.FileMode) {
	v.configPermissions = perm.Perm()
}

// IniLoadOptions sets the load options for ini parsing.
func IniLoadOptions(in ini.LoadOptions) Option {
	return optionFunc(func(v *MConf) {
		v.iniLoadOptions = in
	})
}

func (v *MConf) getConfigType() string {
	if v.configType != "" {
		return v.configType
	}

	cf, err := v.getConfigFile()
	if err != nil {
		return ""
	}

	ext := filepath.Ext(cf)

	if len(ext) > 1 {
		return ext[1:]
	}

	return ""
}

func (v *MConf) getConfigFile() (string, error) {
	if v.configFile == "" {
		return "", errors.New("empty config file")
	}
	return v.configFile, nil
}

// Debug prints all configuration registries for debugging
// purposes.
func Debug() { v.Debug() }

func (v *MConf) Debug() {
	fmt.Printf("Aliases:\n%#v\n", v.aliases)
	fmt.Printf("Override:\n%#v\n", v.override)
	fmt.Printf("Key/Value Store:\n%#v\n", v.kvstore)
	fmt.Printf("Config:\n%#v\n", v.config)
	fmt.Printf("Defaults:\n%#v\n", v.defaults)
}
