package configsecret

import "testing"

func TestSensitiveKeysAreRecognisedHoweverTheyAreSpelled(t *testing.T) {
	// Connector authors name these every way there is, so the match is on
	// substrings and case-insensitive.
	for _, key := range []string{
		"password", "Password", "smtp_password",
		"token", "authToken", "AUTH_TOKEN",
		"apiKey", "api_key", "API-KEY",
		"secret", "clientSecret", "webhook_secret",
		"privateKey", "signature", "credentials",
	} {
		if !IsSensitive(key) {
			t.Errorf("%q was not treated as a secret", key)
		}
	}
}

func TestOrdinarySettingsAreNotMasked(t *testing.T) {
	// The person is on this page to check these. Masking them would make the
	// form useless.
	for _, key := range []string{"url", "channel", "from", "port", "host", "method", "timeout"} {
		if IsSensitive(key) {
			t.Errorf("%q was treated as a secret", key)
		}
	}
}

func TestMaskReplacesSecretsAndKeepsTheRest(t *testing.T) {
	config := map[string]any{
		"url":      "https://hooks.example.com/abc",
		"password": "hunter2",
		"port":     587,
	}
	masked := Mask(config)

	if masked["password"] != Sentinel {
		t.Fatalf("password was returned as %v", masked["password"])
	}
	if masked["url"] != "https://hooks.example.com/abc" {
		t.Fatalf("url was altered: %v", masked["url"])
	}
	if masked["port"] != 587 {
		t.Fatalf("port was altered: %v", masked["port"])
	}
	// The original must be untouched: it may be the map the executor is about
	// to authenticate with.
	if config["password"] != "hunter2" {
		t.Fatal("Mask edited the caller's map in place")
	}
}

func TestAnUnsetSecretStaysEmpty(t *testing.T) {
	// A field nobody has filled in should read as blank, not as a stored
	// credential the person is afraid to touch.
	masked := Mask(map[string]any{"password": ""})
	if masked["password"] != "" {
		t.Fatalf("an empty secret was masked as %v", masked["password"])
	}
}

func TestMergeKeepsWhatWasNotRetyped(t *testing.T) {
	stored := map[string]any{"url": "https://old.example.com", "password": "hunter2"}
	incoming := map[string]any{"url": "https://new.example.com", "password": Sentinel}

	merged := Merge(incoming, stored)
	if merged["password"] != "hunter2" {
		t.Fatalf("the stored password was not kept: %v", merged["password"])
	}
	if merged["url"] != "https://new.example.com" {
		t.Fatalf("the edited url was not applied: %v", merged["url"])
	}
}

func TestMergeAppliesARealReplacement(t *testing.T) {
	merged := Merge(
		map[string]any{"password": "a-new-one"},
		map[string]any{"password": "hunter2"},
	)
	if merged["password"] != "a-new-one" {
		t.Fatalf("a typed replacement was not applied: %v", merged["password"])
	}
}

/*
 * Clearing a credential has to be possible.
 *
 * Treating an empty string as "unchanged" — the obvious shortcut — would make a
 * stored secret impossible to remove through the interface, which is exactly
 * what somebody does when a key leaks.
 */
func TestMergeCanClearASecret(t *testing.T) {
	merged := Merge(
		map[string]any{"password": ""},
		map[string]any{"password": "hunter2"},
	)
	if merged["password"] != "" {
		t.Fatalf("an emptied secret was not cleared: %v", merged["password"])
	}
}

func TestTheSentinelCanNeverBecomeAStoredValue(t *testing.T) {
	// A sentinel for a key that was never stored means nothing; saving it
	// literally would make "__unchanged__" somebody's actual password.
	merged := Merge(map[string]any{"password": Sentinel}, map[string]any{})
	if _, present := merged["password"]; present {
		t.Fatalf("the sentinel was stored: %v", merged["password"])
	}
}

func TestMergeDropsOmittedKeys(t *testing.T) {
	// The form sends the whole configuration, so a key it left out was removed.
	merged := Merge(map[string]any{"url": "x"}, map[string]any{"url": "y", "gone": "z"})
	if _, present := merged["gone"]; present {
		t.Fatal("a removed key survived the merge")
	}
}

func TestHasSentinelSpotsAnUntestableConfig(t *testing.T) {
	if !HasSentinel(map[string]any{"password": Sentinel}) {
		t.Fatal("a placeholder was not spotted")
	}
	if HasSentinel(map[string]any{"password": "real"}) {
		t.Fatal("a real value was reported as a placeholder")
	}
}

func TestNilIsCarriedThrough(t *testing.T) {
	if Mask(nil) != nil {
		t.Fatal("Mask invented a map")
	}
	if Merge(nil, map[string]any{"a": 1}) != nil {
		t.Fatal("Merge invented a map")
	}
}
