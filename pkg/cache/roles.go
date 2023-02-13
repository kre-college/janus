package cache

import (
	"fmt"
	"github.com/hellofresh/janus/pkg/models"
	log "github.com/sirupsen/logrus"
	"sync"
)

type RolesCache struct {
	sync.RWMutex
	Roles map[string]*models.Role
}

func NewRolesCache() *RolesCache {
	roles := make(map[string]*models.Role)

	cache := RolesCache{
		Roles: roles,
	}
	return &cache
}

func (c *RolesCache) Set(role *models.Role) error {
	c.Lock()
	defer c.Unlock()

	c.Roles[role.Name] = &models.Role{
		role.Name,
		role.Features,
	}
	return nil
}

func (c *RolesCache) Get(roleName string) (*models.Role, error) {
	c.RLock()
	defer c.RUnlock()

	role, found := c.Roles[roleName]
	if !found {
		return nil, fmt.Errorf("Cant get %s role", roleName)
	}

	return role, nil
}

func (c *RolesCache) Delete(roleName string) error {
	c.Lock()
	defer c.Unlock()

	if _, found := c.Roles[roleName]; !found {
		log.Printf("Cannot get a %s role", roleName)
	}

	delete(c.Roles, roleName)

	return nil
}

func (c *RolesCache) Update(role *models.Role, roleName string) error {
	c.Lock()
	defer c.Unlock()

	c.Roles[roleName] = &models.Role{
		role.Name,
		role.Features,
	}

	return nil
}
