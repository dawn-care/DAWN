package migrations

import (
	"io/fs"
	"strings"
	"testing"
)

// Deleting a caregiver ends their sessions only through this foreign key's ON DELETE CASCADE;
// DeleteCaregiver never touches "Session". The integration tests check it against Postgres, but
// they skip in CI, so this reads the migrations instead: the last statement in any migration's Up
// that names the constraint must still create it with the cascade.
func TestSessionForeignKeyStillCascadesOnCaregiverDelete(t *testing.T) {
	names, err := fs.Glob(files, "sql/*.sql")
	if err != nil {
		t.Fatal(err)
	}

	var last, lastFile string
	for _, name := range names {
		script, err := fs.ReadFile(files, name)
		if err != nil {
			t.Fatal(err)
		}
		up := strings.SplitN(string(script), "-- +goose Down", 2)[0]
		for _, statement := range strings.Split(up, ";") {
			if strings.Contains(statement, `"Session_user_id_fkey"`) {
				last, lastFile = statement, name
			}
		}
	}

	if last == "" {
		t.Fatal(`no migration creates "Session_user_id_fkey"`)
	}
	for _, want := range []string{"ADD CONSTRAINT", `REFERENCES "Caregiver"`, "ON DELETE CASCADE"} {
		if !strings.Contains(last, want) {
			t.Errorf("%s: the last change to \"Session_user_id_fkey\" no longer has %s, so deleting a caregiver would leave their sessions logged in:\n%s", lastFile, want, strings.TrimSpace(last))
		}
	}
}
