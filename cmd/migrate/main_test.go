package main

import "testing"

func TestMigrationURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dsn     string
		want    string
		wantErr bool
	}{
		{name: "postgres", dsn: "postgres://user:pass@localhost/db", want: "pgx5://user:pass@localhost/db"},
		{name: "postgresql", dsn: "postgresql://user:pass@localhost/db", want: "pgx5://user:pass@localhost/db"},
		{name: "invalid scheme", dsn: "mysql://localhost/db", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := migrationURL(test.dsn)
			if (err != nil) != test.wantErr {
				t.Fatalf("migrationURL() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("migrationURL() = %q, want %q", got, test.want)
			}
		})
	}
}
