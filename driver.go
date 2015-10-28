package brightbox

import (
	"fmt"

	"github.com/brightbox/gobrightbox"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	//	"github.com/docker/machine/libmachine/ssh"
	//"github.com/docker/machine/libmachine/state"
)

const (
	// Docker Machine application client credentials
	defaultClientID     = "app-dkmch"
	defaultClientSecret = "uogoelzgt0nwawb"

	defaultSSHPort = 22
	driverName     = "brightbox"
)

type Driver struct {
	drivers.BaseDriver
	authdetails
	brightbox.ServerOptions
	IPv6       bool
	liveClient *brightbox.Client
}

//Backward compatible Driver factory method
//Using new(brightbox.Driver) is preferred
func NewDriver(hostName, storePath string) Driver {
	return Driver{
		BaseDriver: drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_CLIENT",
			Name:   "brightbox-client",
			Usage:  "Brightbox Cloud API Client",
			Value:  defaultClientID,
		},
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_CLIENT_SECRET",
			Name:   "brightbox-client-secret",
			Usage:  "Brightbox Cloud API Client Secret",
			Value:  defaultClientSecret,
		},
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_USER_NAME",
			Name:   "brightbox-user-name",
			Usage:  "Brightbox Cloud User Name",
		},
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_PASSWORD",
			Name:   "brightbox-password",
			Usage:  "Brightbox Cloud Password for User Name",
		},
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_ACCOUNT",
			Name:   "brightbox-account",
			Usage:  "Brightbox Cloud Account to operate on",
		},
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_REGION",
			Name:   "brightbox-region",
			Usage:  "Brightbox Cloud Region",
			Value:  defaultRegion,
		},
		mcnflag.BoolFlag{
			EnvVar: "BRIGHTBOX_IPV6",
			Name:   "brightbox-ipv6",
			Usage:  "Access server directly over IPv6",
		},
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_ZONE",
			Name:   "brightbox-zone",
			Usage:  "Brightbox Cloud Availability Zone ID",
		},
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_IMAGE",
			Name:   "brightbox-image",
			Usage:  "Brightbox Cloud Image ID",
		},
		mcnflag.StringSliceFlag{
			EnvVar: "BRIGHTBOX_GROUP",
			Name:   "brightbox-group",
			Usage:  "Brightbox Cloud Security Group",
		},
		mcnflag.StringFlag{
			EnvVar: "BRIGHTBOX_TYPE",
			Name:   "brightbox-type",
			Usage:  "Brightbox Cloud Server Type",
		},
	}
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APIClient = flags.String("brightbox-client")
	d.apiSecret = flags.String("brightbox-client-secret")
	d.UserName = flags.String("brightbox-user-name")
	d.password = flags.String("brightbox-password")
	d.Account = flags.String("brightbox-account")
	d.Image = flags.String("brightbox-image")
	d.Region = flags.String("brightbox-region")
	d.ServerType = flags.String("brightbox-type")
	d.IPv6 = flags.Bool("brightbox-ipv6")
	group_list := flags.StringSlice("brightbox-security-group")
	if group_list != nil {
		d.ServerGroups = &group_list
	}
	d.Zone = flags.String("brightbox-zone")
	d.SSHPort = defaultSSHPort
	return d.checkConfig()
}

// Try and avoid authenticating more than once
// Store the authenticated api client in the driver for future use
func (d *Driver) getClient() (*brightbox.Client, error) {
	if d.liveClient != nil {
		log.Debug("Reusing authenticated Brightbox client")
		return d.liveClient, nil
	}
	log.Debug("Authenticating Credentials against Brightbox API")
	client, err := d.authenticatedClient()
	if err == nil {
		d.liveClient = client
		log.Debug("Using authenticated Brightbox client")
	}
	return client, err
}

const (
	errorInvalidRegion        = "Unable to find region name %s"
	errorMandatoryEnvOrOption = "%s must be specified either using the environment variable %s or the CLI option %s"
)

//Statically sanity check flag settings.
func (d *Driver) checkConfig() error {
	if _, ok := regionURL[d.Region]; !ok {
		return fmt.Errorf(errorInvalidRegion, d.Region)
	}
	switch {
	case d.UserName != "" || d.password != "":
		switch {
		case d.UserName == "":
			return fmt.Errorf(errorMandatoryEnvOrOption, "Username", "BRIGHTBOX_USER_NAME", "--brightbox-user-name")
		case d.password == "":
			return fmt.Errorf(errorMandatoryEnvOrOption, "Password", "BRIGHTBOX_PASSWORD", "--brightbox-password")
		}
	case d.APIClient == defaultClientID:
		return fmt.Errorf(errorMandatoryEnvOrOption, "API Client", "BRIGHTBOX_CLIENT", "--brightbox-client")
	}
	return nil
}
