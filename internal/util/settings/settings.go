/*
Package settings - holds configuration specific to a Retrieval Provider node.
*/
package settings

/*
 * Copyright 2020 ConsenSys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"math/big"
	"time"
)

// DefaultTCPInactivityTimeout is the default timeout for TCP inactivity
const DefaultTCPInactivityTimeout = 100 * time.Millisecond

// DefaultLongTCPInactivityTimeout is the default timeout for long TCP inactivity. This timeout should never be ignored.
const DefaultLongTCPInactivityTimeout = 5000 * time.Millisecond

// DefaultRefreshDuration is the default register refresh duration
const DefaultRegisterRefreshDuration = 5000 * time.Millisecond

// AppSettings defines the server configuraiton
type AppSettings struct {
	BindRestAPI    string `mapstructure:"BIND_REST_API"`    // Port number to bind to for client REST API.
	BindGatewayAPI string `mapstructure:"BIND_GATEWAY_API"` // Port number to bind to for gateway TCP communication API.
	BindAdminAPI   string `mapstructure:"BIND_ADMIN_API"`   // Port number to bind to for admin TCP communication API.

	LogLevel      string `mapstructure:"LOG_LEVEL"`       // Log Level: NONE, ERROR, WARN, INFO, TRACE
	LogTarget     string `mapstructure:"LOG_TARGET"`      // Log Level: STDOUT
	LogDir        string `mapstructure:"LOG_DIR"`         // Log Dir: /var/log/fc-retrieval/fc-retrieval-provider
	LogFile       string `mapstructure:"LOG_FILE"`        // Log File: provider.log
	LogMaxBackups int    `mapstructure:"LOG_MAX_BACKUPS"` // Log max backups: 3
	LogMaxAge     int    `mapstructure:"LOG_MAX_AGE"`     // Log max age (days): 28
	LogMaxSize    int    `mapstructure:"LOG_MAX_SIZE"`    // Log max size (MB): 500
	LogCompress   bool   `mapstructure:"LOG_COMPRESS"`    // Log compress: false

	ProviderID     string `mapstructure:"PROVIDER_ID"`      // Node id of this provider
	ProviderSigAlg uint8  `mapstructure:"PROVIDER_SIG_ALG"` // Signature algorithm to be used by private key.

	RegisterAPIURL          string        `mapstructure:"REGISTER_API_URL"`          // Register service url
	RegisterRefreshDuration time.Duration `mapstructure:"REGISTER_REFRESH_DURATION"` // Register refresh duration
	ProviderAddress         string        `mapstructure:"PROVIDER_ADDRESS"`          // Provider address
	ProviderRootSigningKey  string        `mapstructure:"PROVIDER_ROOT_SIGNING_KEY"` // Provider root signing key
	ProviderSigningKey      string        `mapstructure:"PROVIDER_SIGNING_KEY"`      // Provider signing key
	ProviderRegionCode      string        `mapstructure:"PROVIDER_REGION_CODE"`      // Provider region code
	NetworkInfoClient       string        `mapstructure:"NETWORK_CLIENT_INFO"`       // Provider client network info
	NetworkInfoGateway      string        `mapstructure:"NETWORK_GATEWAY_INFO"`      // Provider provider network info
	NetworkInfoAdmin        string        `mapstructure:"NETWORK_ADMIN_INFO"`        // Provider admin network info

	TCPInactivityTimeout     time.Duration `mapstructure:"TCP_INACTIVITY_TIMEOUT"`      // TCP inactivity timeout
	TCPLongInactivityTimeout time.Duration `mapstructure:"TCP_LONG_INACTIVITY_TIMEOUT"` // TCP long inactivity timeout

	SearchPrice *big.Int `mapstructure:SEARCH_PRICE`
	OfferPrice  *big.Int `mapstructure:OFFER_PRICE`
	TopupAmount *big.Int `mapstructure:TOPUP_AMOUNT`
}
