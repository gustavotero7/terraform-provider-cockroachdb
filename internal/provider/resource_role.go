package provider

import (
	"context"
	"fmt"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
	"strings"
)

const (
	attrRoleUsername = "name"
	attrRolePassword = "password"
	attrRoleLogin    = "login"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Roles are groups containing any number of roles and users as members. You can assign privileges to roles, and all members of the role (regardless of whether if they are direct or indirect members) will inherit the role's privileges.",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,

		Schema: map[string]*schema.Schema{
			attrRoleUsername: {
				Description: "Role / User name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			attrRoleLogin: {
				Description: "Allow role to login.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			attrRolePassword: {
				Description: "Role password",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "NULL",
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	username := d.Get(attrRoleUsername).(string)
	login := d.Get(attrRoleLogin).(bool)
	password := d.Get(attrRolePassword).(string)

	loginString := "LOGIN"
	if !login {
		loginString = "NOLOGIN"
	}

	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		query := "CREATE USER " + pq.QuoteIdentifier(username) + " WITH PASSWORD " + pq.QuoteLiteral(password) + " " + loginString
		_, err = tx.Exec(ctx, query)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(username)
	d.Set(attrRoleUsername, username)
	d.Set(attrRoleLogin, login)
	d.Set(attrRolePassword, password)
	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	username := d.Id()
	rows, err := conn.Query(ctx, "SHOW ROLES")
	if err != nil {
		return diag.FromErr(err)
	}

	defer rows.Close()
	for rows.Next() {
		var (
			_username string
			_options  string
		)
		err = rows.Scan(&_username, &_options, nil)
		if err != nil {
			return diag.FromErr(err)
		}

		if username == _username {
			d.SetId(username)
			if err := d.Set(attrRoleUsername, username); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set(attrRoleLogin, !strings.Contains(_options, "NOLOGIN")); err != nil {
				return diag.FromErr(err)
			}
			break
		}
	}

	return nil
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	userName := d.Get(attrRoleUsername).(string)
	login := d.Get(attrRoleLogin).(bool)

	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// update username
		if d.HasChange(attrRoleUsername) {
			oldValue, newValue := d.GetChange(attrRoleUsername)
			oldValueStr := oldValue.(string)
			newValueStr := newValue.(string)
			if newValueStr == "" {
				return fmt.Errorf("username cannot be empty")
			}

			sql := fmt.Sprintf("ALTER ROLE %s RENAME TO %s", pq.QuoteIdentifier(oldValueStr), pq.QuoteIdentifier(newValueStr))
			if _, err := tx.Exec(ctx, sql); err != nil {
				return err
			}
		}
		// update password
		if d.HasChange(attrRolePassword) || d.HasChange(attrRoleUsername) || d.HasChange(attrRoleLogin) {
			loginString := "LOGIN"
			if !login {
				loginString = "NOLOGIN"
			}
			newValue := d.Get(attrRolePassword).(string)
			sql := "ALTER ROLE " + pq.QuoteIdentifier(userName) + " WITH " + loginString + " PASSWORD " + pq.QuoteLiteral(newValue)
			if _, err := tx.Exec(ctx, sql); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(userName)
	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		username := d.Get(attrRoleUsername).(string)
		_, err := tx.Exec(ctx, `DROP USER `+pq.QuoteIdentifier(username))
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
