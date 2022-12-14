package api

// PGRoleAttribute is the attribute for role.
type PGRoleAttribute struct {
	SuperUser   bool `json:"superUser"`
	NoInherit   bool `json:"noInherit"`
	CreateRole  bool `json:"createRole"`
	CreateDB    bool `json:"createDB"`
	CanLogin    bool `json:"canLogin"`
	Replication bool `json:"replication"`
	ByPassRLS   bool `json:"byPassRLS"`
}

// PGRole is the API message for role.
type PGRole struct {
	Name            string           `json:"name"`
	InstanceID      int              `json:"instanceId"`
	ConnectionLimit int              `json:"connectionLimit"`
	ValidUntil      *string          `json:"validUntil"`
	Attribute       *PGRoleAttribute `json:"attribute"`
}

// PGRoleUpsert is the API message for upserting a new role.
type PGRoleUpsert struct {
	Name            string           `json:"name"`
	Password        *string          `json:"password"`
	ConnectionLimit *int             `json:"connectionLimit"`
	ValidUntil      *string          `json:"validUntil"`
	Attribute       *PGRoleAttribute `json:"attribute"`
}
