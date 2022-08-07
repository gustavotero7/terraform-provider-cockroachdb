package provider

import (
	"context"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jackc/pgx/v4"
)

const (
	attrGrantRole = "grant_role"
)

func resourceGrantRole() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Grant role membership.",

		CreateContext: resourceGrantRoleCreate,
		ReadContext:   resourceGrantRoleRead,
		UpdateContext: resourceGrantRoleUpdate,
		DeleteContext: resourceGrantRoleDelete,

		Schema: map[string]*schema.Schema{
			attrRole: {
				Description: "User or role name to grant role to.",
				Type:        schema.TypeString,
				Required:    true,
			},
			attrGrantRole: {
				Description: "The role to grant.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceGrantRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	role := d.Get(attrRole).(string)
	grantRole := d.Get(attrGrantRole).(string)

	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		query := "GRANT " + grantRole + " TO " + role
		_, err = tx.Exec(ctx, query)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildRoleGrantID(role, grantRole))
	return resourceGrantRoleRead(ctx, d, meta)
}

func resourceGrantRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	role := d.Get(attrRole).(string)

	// SHOW GRANTS ON ROLE FOR [role];
	query := "SHOW GRANTS ON ROLE FOR " + role
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return diag.FromErr(err)
	}

	defer rows.Close()
	for rows.Next() {
		var roleName, memberRole string
		err = rows.Scan(&roleName, &memberRole, nil)
		if err != nil {
			return diag.FromErr(err)
		}
		if memberRole == role {
			d.SetId(buildRoleGrantID(memberRole, roleName))
			if err := d.Set(attrGrantRole, roleName); err != nil {
				return diag.FromErr(err)
			}
			break
		}
	}

	return nil
}

func resourceGrantRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceGrantRoleCreate(ctx, d, meta)
}

func resourceGrantRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	role := d.Get(attrRole).(string)
	grantRole := d.Get(attrGrantRole).(string)

	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// REVOKE [grant_role] FROM [role];
		query := "REVOKE " + grantRole + " FROM " + role
		_, err = tx.Exec(ctx, query)
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

func buildRoleGrantID(role, grantRole string) string {
	id := role + "_" + grantRole
	return id
}
