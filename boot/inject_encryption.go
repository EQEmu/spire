package boot

import (
	"github.com/EQEmu/spire/internal/encryption"
	"github.com/google/wire"
)

var encryptionSet = wire.NewSet(
	encryption.NewEncrypter,
)
