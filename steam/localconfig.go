package steam

import (
	"fmt"
	"log/slog"
)

type LocalAppConfig struct {
	LaunchOptions string
}

type LocalConfig struct {
	Apps map[ID]LocalAppConfig
}

func readLocalConfig(fn string) (LocalConfig, error) {
	type Data struct {
		UserLocalConfigStore struct {
			Software struct {
				Valve struct {
					Steam struct {
						Apps map[string]LocalAppConfig
					}
				}
			}
		}
	}
	data, err := ReadVDF[Data](fn)
	if err != nil {
		return LocalConfig{}, nil
	}
	apps := data.UserLocalConfigStore.Software.Valve.Steam.Apps
	cfg := LocalConfig{Apps: make(map[ID]LocalAppConfig, len(apps))}
	for k, v := range apps {
		id, err := ParseID(k)
		if err != nil {
			continue
		}
		cfg.Apps[id] = v
	}
	return cfg, nil
}

func ReadLocalConfig(userID ID) (LocalConfig, error) {
	if userID == 0 {
		return LocalConfig{}, fmt.Errorf("ReadLocalConfig: Invalid userID: %s", userID)
	}
	for _, fn := range FindSteamPaths("userdata", userID.String32(), "config", "localconfig.vdf") {
		cfg, err := readLocalConfig(fn)
		if err == nil {
			return cfg, nil
		}
		Logs.Debug("Cannot read localconfig.vdf", slog.String("fn", fn), slog.String("err", err.Error()))
	}
	return LocalConfig{}, fmt.Errorf("ReadLocalConfig: No localconfig.vdf found for user: %s", userID)
}
