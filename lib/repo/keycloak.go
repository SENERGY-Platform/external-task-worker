/*
 * Copyright 2018 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package repo

import (
	"github.com/SENERGY-Platform/external-task-worker/util"
	"strings"

	"crypto/x509"
	"encoding/base64"

	"time"

	"log"

	"github.com/dgrijalva/jwt-go"
)

type RoleMapping struct {
	Name string `json:"name"`
}

type Keycloak struct {
	config util.Config
}

func (this Keycloak) getUserRoles(user string) (roles []string, err error) {
	clientToken, err := EnsureAccess(this.config)
	if err != nil {
		log.Println("ERROR: getUserRoles::EnsureAccess()", err)
		return roles, err
	}
	roleMappings := []RoleMapping{}
	err = clientToken.GetJSON(this.config.AuthEndpoint+"/auth/admin/realms/master/users/"+user+"/role-mappings/realm", &roleMappings)
	if err != nil {
		log.Println("ERROR: getUserRoles::GetJSON()", err, this.config.AuthEndpoint+"/auth/admin/realms/master/users/"+user+"/role-mappings/realm", string(clientToken))
		return roles, err
	}
	for _, role := range roleMappings {
		roles = append(roles, role.Name)
	}
	return
}

type KeycloakClaims struct {
	RealmAccess RealmAccess `json:"realm_access"`
	jwt.StandardClaims
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}

func (this Keycloak) GetUserToken(user string) (token Impersonate, err error) {
	roles, err := this.getUserRoles(user)
	if err != nil {
		log.Println("ERROR: GetUserToken::getUserRoles()", err)
		return token, err
	}

	// Create the Claims
	claims := KeycloakClaims{
		RealmAccess{Roles: roles},
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(this.config.JwtExpiration)).Unix(),
			Issuer:    this.config.JwtIssuer,
			Subject:   user,
		},
	}

	jwtoken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if this.config.JwtPrivateKey == "" {
		unsignedTokenString, err := jwtoken.SigningString()
		if err != nil {
			log.Println("ERROR: GetUserToken::SigningString()", err)
			return token, err
		}
		tokenString := strings.Join([]string{unsignedTokenString, ""}, ".")
		token = Impersonate("Bearer " + tokenString)
	} else {
		//decode key base64 string to []byte
		b, err := base64.StdEncoding.DecodeString(this.config.JwtPrivateKey)
		if err != nil {
			log.Println("ERROR: GetUserToken::DecodeString()", err)
			return token, err
		}
		//parse []byte key to go struct key (use most common encoding)
		key, err := x509.ParsePKCS1PrivateKey(b)
		tokenString, err := jwtoken.SignedString(key)
		if err != nil {
			log.Println("ERROR: GetUserToken::SignedString()", err)
			return token, err
		}
		token = Impersonate("Bearer " + tokenString)
	}
	return token, err
}
