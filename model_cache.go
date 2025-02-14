package ggm

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

func (m *model[T]) cacheKeyPrefix(id int64) string {
	return fmt.Sprintf("%s-%s-%d", m.connName, m.modelInfo.Name, id)
}

func (m *model[T]) FindBy(id int64) (T, error) {
	t := reflectNew[T]().(T)
	pk := m.pk()
	if pk == "" {
		return t, ErrPrimaryKeyNotDefined
	}
	if cache == nil {
		return m.SelectOne(WhereEq(pk, id))
	}

	key := m.cacheKeyPrefix(id)
	c, err := cache.Get(key)
	if err != nil {
		return t, errors.Wrap(err, "get cache error")
	}
	if c != "" {
		err := json.Unmarshal([]byte(c), t)
		if err != nil {
			return t, err
		}
		return t, nil
	}
	row, err := m.SelectOne(WhereEq(pk, id))
	if err != nil {
		return t, err
	}
	bytes, err := json.Marshal(row)
	if err != nil {
		return t, err
	}
	err = cache.Set(key, string(bytes))
	if err != nil {
		return t, err
	}
	return row, nil
}

func (m *model[T]) UpdateBy(id int64, row T) (int64, error) {
	pk := m.pk()
	if pk == "" {
		return 0, ErrPrimaryKeyNotDefined
	}
	effect, err := m.Update(row, WhereEq(pk, id))
	if err != nil {
		return 0, err
	}
	if cache == nil {
		return effect, nil
	}
	bytes, err := json.Marshal(row)
	if err != nil {
		return 0, err
	}
	key := m.cacheKeyPrefix(id)
	err = cache.Set(key, string(bytes))
	if err != nil {
		return 0, err
	}
	return effect, nil
}
