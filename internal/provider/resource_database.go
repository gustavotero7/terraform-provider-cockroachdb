package provider

import (
	"context"
	"fmt"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/jackc/pgx/v4"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lib/pq"
)

const (
	attrName  = "name"
	attrOwner = "owner"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Database in a CockroachDB cluster.",

		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,

		Schema: map[string]*schema.Schema{
			attrName: {
				Description: "Name of the database.",
				Type:        schema.TypeString,
				Required:    true,
			},
			attrOwner: {
				Description: "Owner of the database.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		name := d.Get(attrName).(string)
		if name == "" {
			return fmt.Errorf("database name can't be an empty string")
		}

		_, err = tx.Exec(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(name))
		if err != nil {
			return err
		}

		var id int
		err = tx.QueryRow(ctx, `SELECT id FROM crdb_internal.databases WHERE name = $1`, name).Scan(
			&id,
		)
		if err != nil {
			return err
		}

		d.SetId(strconv.Itoa(id))
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}
	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	var (
		name  string
		owner string
	)
	err = conn.QueryRow(ctx, `SELECT name, owner FROM crdb_internal.databases WHERE id = $1`, id).Scan(
		&name,
		&owner,
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(attrName, name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(attrOwner, owner); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {

		if d.HasChange(attrName) {
			oldValue, newValue := d.GetChange(attrName)
			oldValueStr := oldValue.(string)
			newValueStr := newValue.(string)
			if newValueStr == "" {
				return fmt.Errorf("database name can't be an empty string")
			}
			_, err := tx.Exec(ctx,
				`ALTER DATABASE `+
					pq.QuoteIdentifier(oldValueStr)+
					` RENAME TO `+
					pq.QuoteIdentifier(newValueStr),
			)
			if err != nil {
				return err
			}

			if err := d.Set(attrName, newValueStr); err != nil {
				return err
			}
		}

		if d.HasChange(attrOwner) {
			dbName := d.Get(attrName).(string)
			_, newValue := d.GetChange(attrOwner)
			newValueStr := newValue.(string)
			_, err := tx.Exec(ctx,
				`ALTER DATABASE `+
					pq.QuoteIdentifier(dbName)+
					` OWNER TO `+
					pq.QuoteIdentifier(newValueStr),
			)
			if err != nil {
				return err
			}

			if err := d.Set(attrOwner, newValueStr); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(*apiClient).Conn(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		name := d.Get(attrName).(string)
		if name == "" {
			return fmt.Errorf("database name can't be an empty string")
		}

		_, err = tx.Exec(ctx, `DROP DATABASE `+pq.QuoteIdentifier(name))
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
