package config

type config_t struct {
    GHEHost   string
	Owner     string
	Repo      string
    CertPath  string
	AppID     int64
    InstallID int64
}

var cur_config *config_t

func NewConfig(GHEHost, Owner, Repo, CertPath string, AppID, InstallID int64) {
    cur_config = &config_t{
        GHEHost,
        Owner,
        Repo,
        CertPath,
        AppID,
        InstallID,
    }
}

func checkIsConfigured() {
    if cur_config == nil {
        panic("GOZBOT not configured")
    }
}

func GHEHost() string {
    checkIsConfigured()
    return cur_config.GHEHost
}

func Owner() string {
    checkIsConfigured()
    return cur_config.Owner
}

func Repo() string {
    checkIsConfigured()
    return cur_config.Repo
}

func CertPath() string {
    checkIsConfigured()
    return cur_config.CertPath
}

func AppID() int64 {
    checkIsConfigured()
    return cur_config.AppID
}

func InstallID() int64 {
    checkIsConfigured()
    return cur_config.InstallID
}
