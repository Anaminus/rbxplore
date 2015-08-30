package settings

import (
	"encoding/json"
	"github.com/anaminus/rbxplore/event"
	"github.com/kardianos/osext"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Settings interface {
	// SetLog specifies an optional logger to output to. Can be nil.
	SetLog(logger *log.Logger)

	// SetFile specifies a file to save to and load from.
	SetFile(file string)

	// SetHook sets up a function to be called after a setting changes. The
	// hook receives the old and new values.
	SetHook(name string, hook func(...interface{})) *event.Connection

	// Load loads the settings from the file, returning whether the operation
	// was successful.
	Load() (ok bool)

	// Save saves the settings to the file, returning whether the operation
	// was successful.
	Save() (ok bool)

	// Get returns the value corresponding to the given name. A nil value is
	// returned if the value does not exist.
	Get(name string) (value interface{})

	// Gets returns a map of every current setting. Modifying this map has no
	// effect on the actual settings.
	Gets() (values map[string]interface{})

	// Set sets the value of a given setting. The new value's type must match
	// the type of the current value. If successful, the settings are saved to
	// the settings file.
	Set(name string, value interface{}) (ok bool)

	// SetFileReloading sets whether settings should be reloaded when the
	// settings file changes.
	SetFileReloading(active bool) error
}

type settingsMap struct {
	defaultName  string
	currentFile  string
	onFileReload event.Event
	log          *log.Logger
	mutex        sync.Mutex
	hooks        map[string]*event.Event
	values       map[string]interface{}
}

func (s *settingsMap) SetLog(logger *log.Logger) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if logger == nil {
		s.log = log.New(ioutil.Discard, "", 0)
		return
	}
	s.log = logger
}

func (s *settingsMap) setFile(file string) {
	if file != "" {
		s.currentFile = file
		return
	}
	exec, err := osext.ExecutableFolder()
	if err != nil {
		log.Printf("could not get location of `%s` file", s.defaultName)
		s.currentFile = ""
		return
	}
	s.currentFile = filepath.Join(exec, s.defaultName)
}

func (s *settingsMap) SetFile(file string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.setFile(file)
}

func (s *settingsMap) SetHook(name string, hook func(...interface{})) *event.Connection {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if hook == nil {
		panic("received nil hook function")
	}
	ev, ok := s.hooks[name]
	if !ok {
		ev = new(event.Event)
		s.hooks[name] = ev
	}
	return ev.Connect(hook)
}

func typesMatch(a, b interface{}) bool {
	switch a.(type) {
	case string:
		if _, ok := b.(string); !ok {
			return false
		}
	case float64:
		if _, ok := b.(float64); !ok {
			return false
		}
	case bool:
		if _, ok := b.(bool); !ok {
			return false
		}
	default:
		return false
	}
	return true
}

func (s *settingsMap) load() (ok bool) {
	f, err := os.Open(s.currentFile)
	if err != nil {
		s.log.Printf("did not load settings file `%s`: %s\n", s.currentFile, err)
		return false
	}
	defer f.Close()
	jd := json.NewDecoder(f)
	m := make(map[string]interface{}, len(s.values))
	if err := jd.Decode(&m); err != nil {
		s.log.Printf("error reading settings file `%s`: %s\n", s.currentFile, err)
		return false
	}
	for name, value := range m {
		old, ok := s.values[name]
		if !ok {
			continue
		}
		if !typesMatch(old, value) {
			s.log.Printf("type mismatch when loading setting `%s`, using current value\n", name)
			continue
		}
		s.values[name] = value
		if hook, ok := s.hooks[name]; ok {
			hook.Fire(old, value)
		}
	}
	return true
}

func (s *settingsMap) save() (ok bool) {
	f, err := os.Create(s.currentFile)
	if err != nil {
		log.Printf("could not open settings file `%s`: %s\n", s.currentFile, err)
		return false
	}
	defer f.Close()

	b, err := json.MarshalIndent(&s.values, "", "\t")
	if err != nil {
		log.Printf("error encoding settings file `%s`: %s\n", s.currentFile, err)
		return false
	}

	if _, err := f.Write(b); err != nil {
		log.Printf("error writing settings file `%s`: %s\n", s.currentFile, err)
		return false
	}

	return true
}

func (s *settingsMap) Load() (ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.load()
}

func (s *settingsMap) Save() (ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.save()
}

func (s *settingsMap) Get(name string) (value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.values[name]
}

func (s *settingsMap) Gets() (values map[string]interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	values = make(map[string]interface{}, len(s.values))
	for name, value := range s.values {
		values[name] = value
	}
	return values
}

func (s *settingsMap) Set(name string, value interface{}) (ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	old, ok := s.values[name]
	if !ok {
		return false
	}
	if !typesMatch(old, value) {
		return false
	}
	s.values[name] = value
	s.save()

	if hook, ok := s.hooks[name]; ok {
		hook.Fire(old, value)
	}
	return true
}

func (s *settingsMap) SetFileReloading(active bool) error {
	panic("not implemented")
	return nil
}

// Create creates a new Settings object. Initial settings are specified with
// the initialValues map. Note that settings cannot be added or removed after
// this point, only changed. Only string, float64, and bool types are allowed.
//
// defaultName specifies the name of a file, which will be in the same
// location as the executable. This file is used when no file has been
// specified with SetFile.
func Create(defaultName string, initialValues map[string]interface{}) Settings {
	s := &settingsMap{
		defaultName: defaultName,
		log:         log.New(ioutil.Discard, "", 0),
		hooks:       make(map[string]*event.Event, len(initialValues)),
	}
	s.setFile("")
	s.values = make(map[string]interface{}, len(initialValues))
	for name, value := range initialValues {
		switch value.(type) {
		case string, float64, bool:
			s.values[name] = value
		default:
			panic("invalid setting type for `" + name + "` (must be either string, float64, or bool")
		}
	}
	return s
}
