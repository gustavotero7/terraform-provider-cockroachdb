package provider

import (
	"context"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jackc/pgx/v4"
	"strings"
)

const (
	attrRole       = "role"
	attrObjectType = "object_type"
	attrObjects    = "objects"
	attrPrivileges = "privileges"
)

func resourceGrant() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "The GRANT statement controls each role or user's SQL privileges for interacting with specific databases, schemas, tables, or user-defined types.",

		CreateContext: resourceGrantCreate,
		ReadContext:   resourceGrantRead,
		UpdateContext: resourceGrantUpdate,
		DeleteContext: resourceGrantDelete,

		Schema: map[string]*schema.Schema{
			attrRole: {
				Description: "Target role Name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			attrObjectType: {
				Description: "Object type. Must be one of the following: database, schema, table.",
				Type:        schema.TypeString,
				Required:    true,
			},
			attrObjects: {
				Description: "Objects to grant privileges on.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			attrPrivileges: {
				Description: "Privileges to grant.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
		},
	}
}

func resourceGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	role := d.Get(attrRole).(string)
	objectType := d.Get(attrObjectType).(string)
	objects := d.Get(attrObjects).([]interface{})
	privileges := d.Get(attrPrivileges).([]interface{})
	objectsStr := sliceInterfacesToStrings(objects)
	privilegesStr := sliceInterfacesToStrings(privileges)

	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		query := "GRANT " + strings.Join(privilegesStr, ", ") + " ON " + objectType + " " + strings.Join(objectsStr, ",") + " TO " + role
		_, err = tx.Exec(ctx, query)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildGrantID(role, objectType))
	d.Set(attrRole, role)
	d.Set(attrObjectType, objectType)
	d.Set(attrObjects, objects)
	d.Set(attrPrivileges, privileges)
	return resourceGrantRead(ctx, d, meta)
}

func resourceGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	role := d.Get(attrRole).(string)
	objectType := d.Get(attrObjectType).(string)
	objects := sliceInterfacesToStrings(d.Get(attrObjects).([]interface{}))
	statePrivileges := sliceInterfacesToStrings(d.Get(attrPrivileges).([]interface{}))

	query := "SHOW GRANTS ON " + objectType + " " + strings.Join(objects, ",") + " FOR " + role
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return diag.FromErr(err)
	}

	privilegesMap := make(map[string]struct{})
	privilegeTypeIndex := -1
	defer rows.Close()
	for rows.Next() {
		if privilegeTypeIndex == -1 {
			for i, description := range rows.FieldDescriptions() {
				if string(description.Name) == "privilege_type" {
					privilegeTypeIndex = i
					break
				}
			}
		}

		if privilegeTypeIndex == -1 {
			return diag.Errorf("failed to infer privilege_type column index")
		}

		vales, err := rows.Values()
		if err != nil {
			return diag.FromErr(err)
		}
		privilege, _ := vales[privilegeTypeIndex].(string)
		privilegesMap[privilege] = struct{}{}
	}

	// reuse state privileges if it contains all the items in privilegesMap
	samesAsState := len(statePrivileges) == len(privilegesMap)
	if samesAsState {
		for _, privilege := range statePrivileges {
			if _, ok := privilegesMap[privilege]; !ok {
				samesAsState = false
				break
			}
		}
	}

	if !samesAsState {
		statePrivileges = make([]string, 0)
		for s, _ := range privilegesMap {
			statePrivileges = append(statePrivileges, s)
		}
	}

	d.SetId(buildGrantID(role, objectType))
	if err := d.Set(attrRole, role); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(attrObjectType, objectType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(attrObjects, objects); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(attrPrivileges, statePrivileges); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceGrantUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if res := resourceGrantDelete(ctx, d, meta); res.HasError() {
		return res
	}
	return resourceGrantCreate(ctx, d, meta)
}

func resourceGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	role := d.Get(attrRole).(string)
	objectType := d.Get(attrObjectType).(string)
	objects := d.Get(attrObjects).([]interface{})
	privileges := d.Get(attrPrivileges).([]interface{})
	objectsStr := sliceInterfacesToStrings(objects)
	privilegesStr := sliceInterfacesToStrings(privileges)

	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		query := "REVOKE " + strings.Join(privilegesStr, ", ") + " ON " + objectType + " " + strings.Join(objectsStr, ",") + " FROM " + role
		_, err := tx.Exec(ctx, query)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func buildGrantID(role, objectType string) string {
	id := role + "_" + objectType
	return id
}
