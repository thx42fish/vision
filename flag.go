package vision

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"
)

// Commandline args -> Environment args -> Config properties.
// Command line example: -server-addr=localhost
// Environment example: SERVER_ADDR=localhost
// Config file: {"server.addr":"localhost"} or server.addr=localhost
type Parser struct {
	flagSet      *flag.FlagSet
	envEnable    bool
	envPrefix    string
	filename     string
	fileFlagName string // config file flag name in command line, some like `c`
	ignoreKeys   []string
	parsed       bool
	mu           sync.Mutex
}

type Option func(sf *Parser)

func New(options ...Option) *Parser {
	sf := new(Parser)
	for _, op := range options {
		op(sf)
	}
	return sf
}

func WithFlagSet(flagSet *flag.FlagSet) Option {
	return func(sf *Parser) {
		sf.flagSet = flagSet
	}
}

func WithFlagIgnore(ignoreKeys []string) Option {
	return func(sf *Parser) {
		sf.ignoreKeys = ignoreKeys
	}
}

// config file flag name in command line, some like `c`
func WithFlagFile(fileFlagName string) Option {
	return func(sf *Parser) {
		sf.fileFlagName = fileFlagName
	}
}

func WithEnvEnable() Option {
	return func(sf *Parser) {
		sf.envEnable = true
	}
}

func WithEnvPrefix(prefix string) Option {
	return func(sf *Parser) {
		sf.envPrefix = prefix
	}
}

func (s *Parser) Parse() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.parsed {
		return
	}
	if s.flagSet == nil {
		s.flagSet = flag.CommandLine
	}
	if !s.flagSet.Parsed() {
		err = s.flagSet.Parse(os.Args[1:])
		if err != nil {
			return
		}
		s.parsed = true
	}
	// collect set args.
	sets := make(map[string]int)
	s.flagSet.Visit(func(f *flag.Flag) {
		sets[f.Name] = 0
	})

	// try find properties in envEnable vars.
	// upper words required, example: TEMP_DIR
	if s.envEnable {
		s.replaceByENV(sets, s.envPrefix)
	}

	// try find by config file flag name.
	if s.fileFlagName != "" {
		var filename string
		s.flagSet.Visit(func(f *flag.Flag) {
			if f.Name == s.fileFlagName {
				filename = f.Value.String()
			}
		})
		if filename != "" {
			filename = HomeAbs(filename)
			if _, innerErr := os.Stat(filename); innerErr == nil || innerErr == os.ErrExist {
				s.filename = filename
			}
		}
	}
	// try to find the config file.
	if s.filename != "" {
		// ignore wrong file given.
		if _, innerErr := os.Stat(s.filename); innerErr != nil && innerErr != os.ErrExist {
			return
		}
		err = s.replaceByFile(sets, s.filename)
		if err != nil {
			return
		}
	}
	return
}

func (s *Parser) replaceByFile(sets map[string]int, file string) (err error) {
	_, err = os.Stat(file)
	if err != nil && err != os.ErrExist {
		err = fmt.Errorf("Invalid config filename path given, filename: %s ", file)
		return
	}
	viper.SetConfigFile(file)
	err = viper.ReadInConfig()
	if err == nil {
		s.flagSet.VisitAll(func(f *flag.Flag) {
			// replace by config.
			if !s.ignore(sets, f.Name) {
				key := strings.ReplaceAll(f.Name, "-", ".")
				if viper.IsSet(key) {
					val := viper.GetString(key)
					_ = s.flagSet.Set(f.Name, val)
					sets[f.Name] = 0
				}
			}
		})
		return
	}
	err = fmt.Errorf("Read config filename failed, filename: %s, error: %v ", file, err)
	return
}

func (s *Parser) replaceByENV(sets map[string]int, envPrefix string) {
	s.flagSet.VisitAll(func(f *flag.Flag) {
		// replace by envEnable.
		if !s.ignore(sets, f.Name) {
			var val string
			var exist bool
			env := strings.ReplaceAll(f.Name, "-", "_")
			if envPrefix != "" {
				env = strings.Join([]string{envPrefix, env}, "")
			}
			env = strings.ToUpper(env)
			val, exist = os.LookupEnv(env)
			if exist {
				_ = s.flagSet.Set(f.Name, val)
				sets[f.Name] = 0
			}
		}
	})
	return
}

func (s *Parser) ignore(sets map[string]int, key string) bool {
	if _, exist := sets[key]; exist {
		return true
	}
	if len(s.ignoreKeys) > 0 {
		for _, v := range s.ignoreKeys {
			if v == key {
				return true
			}
		}
	}
	return false
}
