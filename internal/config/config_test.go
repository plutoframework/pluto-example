package config

import "testing"

func lookup(values map[string]string) Lookup {
	return func(key string) (string, bool) { value, ok := values[key]; return value, ok }
}

func TestLoadAndValidate(t *testing.T) {
	defaults := map[string]string{Prefix + "DEMO_AUTH": "true"}

	settings, err := LoadFrom(lookup(defaults))
	if err != nil {
		t.Fatal(err)
	}

	if settings.Features.PurchaseExport != nil ||
		settings.Database.Mode != "memory" ||
		!settings.Server.Gzip {
		t.Fatalf("bad defaults: %+v", settings)
	}

	defaults[Prefix+"FEATURE_PURCHASE_EXPORT"] = "false"
	defaults[Prefix+"SERVER_GZIP"] = "false"

	settings, err = LoadFrom(lookup(defaults))
	if err != nil || settings.Features.PurchaseExport == nil ||
		*settings.Features.PurchaseExport ||
		settings.Server.Gzip {
		t.Fatalf("explicit false lost: %+v %v", settings, err)
	}

	for _, test := range []struct{ name, value string }{
		{
			"DEMO_AUTH",
			"false",
		},
		{
			"DEMO_AUTH",
			"maybe",
		},
		{
			"FEATURE_PURCHASE_EXPORT",
			"sometimes",
		},
		{
			"SERVER_GZIP",
			"yes",
		},
		{
			"SERVER_ADDRESS",
			"0.0.0.0:8080",
		},
		{
			"SERVER_ADDRESS",
			":8080",
		},
		{
			"DATABASE_MODE",
			"sqlite",
		},
		{
			"DATABASE_MODE",
			"postgres",
		},
		{
			"LOG_LEVEL",
			"verbose",
		},
	} {
		t.Run(test.name+test.value, func(t *testing.T) {
			values := map[string]string{
				Prefix + "DEMO_AUTH": "true",
				Prefix + test.name:   test.value,
			}
			if _, err := LoadFrom(lookup(values)); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}

	if _, err := LoadFrom(lookup(nil)); err == nil {
		t.Fatal("demo auth enabled by default")
	}
}
