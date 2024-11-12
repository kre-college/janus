package authorization

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID               uint64               `json:"UserID"`
	FullUserName         string               `json:"FullUserName"`
	Roles                RolesToID            `json:"Roles"`
	CurrentInstitutionID uint64               `json:"CurrentInstitutionID"`
	InstitutionIDToRoles map[uint64]RolesToID `json:"InstitutionIDToRoles"`
	jwt.StandardClaims
}

type RoleToID struct {
	Role string  `json:"role"`
	ID   *uint64 `json:"id"`
}

type RolesToID []*RoleToID

func (rti RolesToID) GetRoles() []string {
	roles := make([]string, len(rti))
	for i, r := range rti {
		roles[i] = r.Role
	}

	return roles
}

func ExtractClaims(jwtToken string) (*Claims, error) {
	parts := strings.Split(jwtToken, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("JWT token invalid")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	err = json.Unmarshal(decoded, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
