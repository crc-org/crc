package config

type InMemoryStorage struct {
	storage map[string]interface{}
}

func NewEmptyInMemoryStorage() *InMemoryStorage {
	return NewInMemoryStorage(make(map[string]interface{}))
}

func NewInMemoryStorage(init map[string]interface{}) *InMemoryStorage {
	return &InMemoryStorage{
		storage: init,
	}
}

func (s *InMemoryStorage) Get(key string) interface{} {
	return s.storage[key]
}

func (s *InMemoryStorage) Set(key string, value interface{}) error {
	s.storage[key] = value
	return nil
}

func (s *InMemoryStorage) Unset(key string) error {
	delete(s.storage, key)
	return nil
}
