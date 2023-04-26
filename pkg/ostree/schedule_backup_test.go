package ostree

import (
	"fmt"
	"testing"
)

var ostreeStatusJSON = `{
  "deployments" : [
    {
      "requested-packages" : [ "exa", "greenboot", "langpacks-en", "neovim", "vim" ],
      "requested-base-local-replacements" : [],
      "pending-base-timestamp" : 1682197278,
      "requested-modules" : [],
      "signatures" : [ [ true, false, false, false, false, "6A51BBABBA3D5467B6171221809A8D7CEB10B464", 1681459754, 0, "RSA", "SHA256", "Fedora", "fedora-38-primary@fedoraproject.org", "6A51BBABBA3D5467B6171221809A8D7CEB10B464", 0, 0 ] ],
      "regenerate-initramfs" : false,
      "pending-base-version" : "38.20230422.1",
      "version" : "38.20230414.n.0",
      "requested-local-fileoverride-packages" : [],
      "base-commit-meta" : {
        "ostree.linux" : "6.2.9-300.fc38.x86_64",
        "version" : "38.20230414.n.0",
        "ostree.bootable" : true,
        "rpmostree.inputhash" : "613038af0e67dfebb816b61b960bcb7e4f137050908e84f9ba726a4914aab319",
        "rpmostree.rpmmd-repos" : []
      },
      "base-remote-replacements" : {      },
      "layered-commit-meta" : {
        "rpmostree.clientlayer" : true,
        "version" : "38.20230414.n.0",
        "rpmostree.clientlayer_version" : 6,
        "rpmostree.state-sha512" : "7943522812e91e002bfc9f4abbf9d885d5dce3026ea3e6a16f7f43db26080b40d171d478fd9feb9d24120311f503104665e2bfca89e407e0ec25d073059cf764",
        "rpmostree.rpmmd-repos" : [],
        "ostree.linux" : "6.2.9-300.fc38.x86_64",
        "ostree.bootable" : true
      },
      "timestamp" : 1682504789,
      "packages" : [ "exa", "greenboot", "langpacks-en", "neovim", "vim" ],
      "base-local-replacements" : [],
      "osname" : "fedora",
      "pending-base-checksum" : "233d5d86c58d4da70da4e1aec1c457c0b1b4a66fd5544d103d86a0280956e09d",
      "pinned" : false,
      "requested-modules-enabled" : [],
      "modules" : [],
      "booted" : false,
      "base-removals" : [],
      "unlocked" : "none",
      "requested-base-removals" : [],
      "base-checksum" : "7f13707d7180ddf167cf796fee7e1d3238fc20517a6cc3aa03108eb2d325a467",
      "id" : "fedora-8523da044806b21d28963a07c3e21e01d9a00dfeb76d4dbcdb582e8b3bf4d7b2.0",
      "origin" : "fedora:fedora/38/x86_64/silverblue",
      "serial" : 0,
      "base-timestamp" : 1681459461,
      "gpg-enabled" : true,
      "base-version" : "38.20230414.n.0",
      "requested-local-packages" : [],
      "checksum" : "8523da044806b21d28963a07c3e21e01d9a00dfeb76d4dbcdb582e8b3bf4d7b2",
      "staged" : true
    },
    {
      "requested-packages" : [ "greenboot", "langpacks-en", "neovim", "vim" ],
      "requested-base-local-replacements" : [],
      "pending-base-timestamp" : 1682197278,
      "requested-modules" : [],
      "signatures" : [ [ true, false, false, false, false, "6A51BBABBA3D5467B6171221809A8D7CEB10B464", 1681459754, 0, "RSA", "SHA256", "Fedora", "fedora-38-primary@fedoraproject.org", "6A51BBABBA3D5467B6171221809A8D7CEB10B464", 0, 0 ] ],
      "regenerate-initramfs" : false,
      "pending-base-version" : "38.20230422.1",
      "version" : "38.20230414.n.0",
      "requested-local-fileoverride-packages" : [],
      "base-commit-meta" : {
        "ostree.linux" : "6.2.9-300.fc38.x86_64",
        "version" : "38.20230414.n.0",
        "ostree.bootable" : true,
        "rpmostree.inputhash" : "613038af0e67dfebb816b61b960bcb7e4f137050908e84f9ba726a4914aab319",
        "rpmostree.rpmmd-repos" : []
      },
      "base-remote-replacements" : {      },
      "layered-commit-meta" : {
        "rpmostree.clientlayer" : true,
        "version" : "38.20230414.n.0",
        "rpmostree.clientlayer_version" : 6,
        "rpmostree.state-sha512" : "ee5f714942a37a467188128c0813a2caf7a9ceb7a9747122bfe870d6917b9085be553fd8c02fcfcf379bb8064aba9631051209e2cb7595e81b8246d8b39e9ba1",
        "rpmostree.rpmmd-repos" : [],
        "ostree.linux" : "6.2.9-300.fc38.x86_64",
        "ostree.bootable" : true
      },
      "timestamp" : 1682247419,
      "packages" : [ "greenboot", "langpacks-en", "neovim", "vim" ],
      "base-local-replacements" : [],
      "osname" : "fedora",
      "pending-base-checksum" : "233d5d86c58d4da70da4e1aec1c457c0b1b4a66fd5544d103d86a0280956e09d",
      "pinned" : false,
      "requested-modules-enabled" : [],
      "modules" : [],
      "booted" : true,
      "base-removals" : [],
      "unlocked" : "none",
      "requested-base-removals" : [],
      "base-checksum" : "7f13707d7180ddf167cf796fee7e1d3238fc20517a6cc3aa03108eb2d325a467",
      "id" : "fedora-551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627.0",
      "origin" : "fedora:fedora/38/x86_64/silverblue",
      "serial" : 0,
      "base-timestamp" : 1681459461,
      "gpg-enabled" : true,
      "base-version" : "38.20230414.n.0",
      "requested-local-packages" : [],
      "checksum" : "551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627",
      "staged" : false
    },
    {
      "requested-packages" : [ "bat", "greenboot", "langpacks-en", "neovim", "tmux", "vim" ],
      "requested-base-local-replacements" : [],
      "pending-base-timestamp" : 1682197278,
      "requested-modules" : [],
      "signatures" : [ [ true, false, false, false, false, "6A51BBABBA3D5467B6171221809A8D7CEB10B464", 1681459754, 0, "RSA", "SHA256", "Fedora", "fedora-38-primary@fedoraproject.org", "6A51BBABBA3D5467B6171221809A8D7CEB10B464", 0, 0 ] ],
      "regenerate-initramfs" : false,
      "pending-base-version" : "38.20230422.1",
      "version" : "38.20230414.n.0",
      "requested-local-fileoverride-packages" : [],
      "base-commit-meta" : {
        "ostree.linux" : "6.2.9-300.fc38.x86_64",
        "version" : "38.20230414.n.0",
        "ostree.bootable" : true,
        "rpmostree.inputhash" : "613038af0e67dfebb816b61b960bcb7e4f137050908e84f9ba726a4914aab319",
        "rpmostree.rpmmd-repos" : []
      },
      "base-remote-replacements" : {      },
      "layered-commit-meta" : {
        "rpmostree.clientlayer" : true,
        "version" : "38.20230414.n.0",
        "rpmostree.clientlayer_version" : 6,
        "rpmostree.state-sha512" : "a042a2bb9b453a335c4d30863f54d92d6509cd27ef687ecdd6b9b43121ee8c0aff1e341d445868e404b1089f4ae5273a284f0e5c10baef57e4fadc3efb9e5aa5",
        "rpmostree.rpmmd-repos" : [],
        "ostree.linux" : "6.2.9-300.fc38.x86_64",
        "ostree.bootable" : true
      },
      "timestamp" : 1682329250,
      "packages" : [ "bat", "greenboot", "langpacks-en", "neovim", "tmux", "vim" ],
      "base-local-replacements" : [],
      "osname" : "fedora",
      "pending-base-checksum" : "233d5d86c58d4da70da4e1aec1c457c0b1b4a66fd5544d103d86a0280956e09d",
      "pinned" : false,
      "requested-modules-enabled" : [],
      "modules" : [],
      "booted" : false,
      "base-removals" : [],
      "unlocked" : "none",
      "requested-base-removals" : [],
      "base-checksum" : "7f13707d7180ddf167cf796fee7e1d3238fc20517a6cc3aa03108eb2d325a467",
      "id" : "fedora-d1846d3f6fa89e6967cc740313ea727126e02db5b7af103f171d0e8f432a3d4d.0",
      "origin" : "fedora:fedora/38/x86_64/silverblue",
      "serial" : 0,
      "base-timestamp" : 1681459461,
      "gpg-enabled" : true,
      "base-version" : "38.20230414.n.0",
      "requested-local-packages" : [],
      "checksum" : "d1846d3f6fa89e6967cc740313ea727126e02db5b7af103f171d0e8f432a3d4d",
      "staged" : false
    }
  ],
  "transaction" : null,
  "cached-update" : null,
  "update-driver" : null
}`

func Test_getOstreeStatus(t *testing.T) {
	expectedStatus := ostreeStatus{
		Deployments: []ostreeDeployment{
			{
				ID:     "fedora-8523da044806b21d28963a07c3e21e01d9a00dfeb76d4dbcdb582e8b3bf4d7b2.0",
				Booted: false,
			},
			{
				ID:     "fedora-551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627.0",
				Booted: true,
			},
			{
				ID:     "fedora-d1846d3f6fa89e6967cc740313ea727126e02db5b7af103f171d0e8f432a3d4d.0",
				Booted: false,
			},
		},
	}

	runOstreeStatus = func() ([]byte, error) {
		return []byte(ostreeStatusJSON), nil
	}

	status, err := getOstreeStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(status.Deployments) != len(expectedStatus.Deployments) {
		t.Fatalf("wrong number of deployments")
	}

	for _, d := range status.Deployments {
		matching, found := func() (ostreeDeployment, bool) {
			for _, ed := range expectedStatus.Deployments {
				if d.ID == ed.ID {
					return ed, true
				}
			}
			return ostreeDeployment{}, false
		}()

		if !found {
			t.Fatalf("deployment %s not found", d.ID)
		}

		if matching.Booted != d.Booted {
			t.Fatalf("deployment %s's Booted doesn't match (got: %v, expected: %v)",
				d.ID, d.Booted, matching.Booted)
		}
	}
}

func Test_getOstreeStatus_ostreeFail(t *testing.T) {
	runOstreeStatus = func() ([]byte, error) {
		return []byte{}, fmt.Errorf("some error")
	}

	_, err := getOstreeStatus()
	if err == nil {
		t.Fatalf("expected an error to happen")
	}
}

func Test_getBootedOstreeID(t *testing.T) {
	tests := []struct {
		name       string
		input      ostreeStatus
		shouldErr  bool
		expectedID string
	}{
		{
			name:      "0 deployments",
			input:     ostreeStatus{Deployments: []ostreeDeployment{}},
			shouldErr: true,
		},
		{
			name: "1 deployment: booted",
			input: ostreeStatus{Deployments: []ostreeDeployment{
				{
					ID:     "fedora-551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627.0",
					Booted: true,
				},
			}},
			expectedID: "fedora-551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627.0",
		},
		{
			name: "2 deployments: booted and rollback",
			input: ostreeStatus{Deployments: []ostreeDeployment{
				{
					ID:     "fedora-551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627.0",
					Booted: true,
				},
				{
					ID:     "fedora-d1846d3f6fa89e6967cc740313ea727126e02db5b7af103f171d0e8f432a3d4d.0",
					Booted: false,
				},
			}},
			expectedID: "fedora-551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627.0",
		},
		{
			name: "3 deployments: staged, booted, and rollback",
			input: ostreeStatus{Deployments: []ostreeDeployment{
				{
					ID:     "fedora-8523da044806b21d28963a07c3e21e01d9a00dfeb76d4dbcdb582e8b3bf4d7b2.0",
					Booted: false,
				},
				{
					ID:     "fedora-551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627.0",
					Booted: true,
				},
				{
					ID:     "fedora-d1846d3f6fa89e6967cc740313ea727126e02db5b7af103f171d0e8f432a3d4d.0",
					Booted: false,
				},
			}},
			expectedID: "fedora-551647113bff576a61185a056d7d2359751eac87375f2668955805e404cc6627.0",
		},
		{
			name: "2 deployments but none is booted",
			input: ostreeStatus{Deployments: []ostreeDeployment{
				{
					ID:     "fedora-8523da044806b21d28963a07c3e21e01d9a00dfeb76d4dbcdb582e8b3bf4d7b2.0",
					Booted: false,
				},
				{
					ID:     "fedora-d1846d3f6fa89e6967cc740313ea727126e02db5b7af103f171d0e8f432a3d4d.0",
					Booted: false,
				},
			}},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := getBootedOstreeID(tt.input)
			if tt.shouldErr && err == nil {
				t.Fatalf("expected an error, but didn't happen")
			} else if !tt.shouldErr && err != nil {
				t.Fatalf("unexpected error")
			}

			if id != tt.expectedID {
				t.Fatalf("got wrong ID (%s), expected %s", id, tt.expectedID)
			}
		})
	}
}
