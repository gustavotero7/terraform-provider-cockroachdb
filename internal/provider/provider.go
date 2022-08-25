package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jackc/pgx/v4"
	"net/url"
	"strings"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"host": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "CockroachDB server address",
				},
				"port": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "CockroachDB server port",
					Default:     "26257",
				},
				"username": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "CockroachDB username to connect with.",
					Sensitive:   true,
				},
				"password": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "CockroachDB password to connect with.",
					Sensitive:   true,
				},
				"database": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "CockroachDB database name to connect to.",
				},
				"sslmode": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "CockroachDB SSL mode to use.",
					Default:     "verify-full",
				},
				"cluster": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "CockroachDB cluster id.",
				},
			},

			DataSourcesMap: map[string]*schema.Resource{},
			ResourcesMap: map[string]*schema.Resource{
				"cockroachdb_database":   resourceDatabase(),
				"cockroachdb_role":       resourceRole(),
				"cockroachdb_grant":      resourceGrant(),
				"cockroachdb_grant_role": resourceGrantRole(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)
		return p
	}
}

type apiClient struct {
	dsn string
}

func (c *apiClient) Conn(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, c.dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}
	return conn, nil
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		host := d.Get("host").(string)
		port := d.Get("port").(int)
		database := d.Get("database").(string)
		username := d.Get("username").(string)
		password := url.PathEscape(d.Get("password").(string))
		sslmode := d.Get("sslmode").(string)
		cluster := d.Get("cluster").(string)

		dsn := "postgresql://" + username + ":" + password + "@" + host + ":" + fmt.Sprint(port) + "/" + database + "?sslmode=" + sslmode
		if cluster != "" {
			dsn += "&options=--cluster%3D" + cluster
		}

		//// Use dsn from env if set and host config is empty. (used for integration tests)
		//if envDSN := os.Getenv("COCKROACHDB_DSN"); host == "" && envDSN != "" {
		//	dsn = envDSN
		//}

		return &apiClient{dsn: dsn}, diag.Diagnostics{}
	}
}
