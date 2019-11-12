package main

import (
	"encoding/json"
	// "fmt"
	"io/ioutil"
	// "io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	// "syscall"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// Config : Global Configuration
type Config struct {
	InstallParameters *InstallParameters `json:"install_parameters"`
	SheetConfig       *SheetConfig       `json:"google_sheet_config"`
	BinaryParameters  *BinaryParameters  `json:"binary_parameters"`
	KalturaSettings   *KalturaSettings   `json:"kaltura_classroomn_localsettings"`
}

// InstallParameters : PostInstall Cconfiguration settings
type InstallParameters struct {
	Silent          string `json:"silent"`
	InstallDir      string `json:"install_dir"`
	RecordingDir    string `json:"recording_dir"`
	URL             string `json:"url"`
	AppToken        string `json:"apptoken"`
	AppTokenID      string `json:"apptoken_id"`
	PartnerID       string `json:"partner_id"`
	DesktopShortcut string `json:"desktop_shortcut"`
	ProgramShortcut string `json:"program_shortcut"`
}

// BinaryParameters : Parameters to download kaltura binary
type BinaryParameters struct {
	URL          string `json:"url"`
	Checksum     string `json:"checksum"`
	FileLocation string `json:"file_location"`
}

// SheetConfig : Configuration for google sheet
type SheetConfig struct {
	Env           string `json:"env"`
	SpreadsheetID string `json:"speadsheet_id"`
	Scopes        string `json:"scopes"`
	SheetRange    string `json:"range"`
}

// KalturaSettings : Kaltura Classroom local settings
type KalturaSettings struct {
	ResourceID   string `json:"resourceID"`
	LaunchSilent string `json:"luanch_silent"`
	Countdown    string `json:"countdown"`
}

// Grabs the kaltura configuration file
func getKalturaConfig(path string) map[string]interface{} {

	var kaltura map[string]interface{}

	// Open our jsonFile
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[INFO] successfully opened localSettings")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// json file's content into 'config' which we defined above
	json.Unmarshal(byteValue, &kaltura)

	return kaltura
}

func updateKalturaSettings(path string, newSettings map[string]interface{}) error {

	marshalledSettings, _ := json.MarshalIndent(newSettings, "", "\t")
	err := ioutil.WriteFile(path, marshalledSettings, 0644)

	if err != nil {
		return err
	}

	return nil
}

func getConfig() (*Config, error) {

	var config Config

	// Open our jsonFile
	jsonFile, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}

	log.Println("[INFO] successfully opened config.json")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'config' which we defined above
	json.Unmarshal(byteValue, &config)

	return &config, nil
}

// func installMSI(binParams *BinaryParameters, installParams *InstallParameters) error {

// 	// Download Binary
// 	err := downloadFile(binParams.FileLocation, binParams.URL)
// 	if err != nil {
// 		return err
// 	}

// 	tmplString := `/i "%s" /qb /norestart ADDLOCAL=ALL KALTURA_URL=%s KALTURA_APPTOKEN=%s KALTURA_APPTOKEN_ID=%s KALTURA_PARTNER_ID=%s INSTALLDESKTOPSHORTCUT=%s INSTALLPROGRAMSSHORTCUT=%s /L*V "C:\VCU-Deploy\logs\Kaltura-Classroom-Install.log"`

// 	installString := fmt.Sprintf(tmplString,
// 		binParams.FileLocation,
// 		// installParams.InstallDir,
// 		// installParams.RecordingDir,
// 		installParams.URL,
// 		installParams.AppToken,
// 		installParams.AppTokenID,
// 		installParams.PartnerID,
// 		installParams.DesktopShortcut,
// 		installParams.ProgramShortcut,
// 	)

// 	log.Println("[INFO] msiexec.exe " + installString)

// 	// Put string in powershell file
// 	out, err := os.Create("msiInstall.ps1")
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()

// 	// Write the body to file
// 	out.WriteString("msiexec.exe " + installString)

// 	cmd := exec.Command("powershell.exe", "msiInstall.ps1")
// 	if err = cmd.Run(); err != nil {
// 		log.Println("[ERROR] could not install kaltura")
// 		return err
// 	}

// 	return nil
// }

func generateKalturaConfig(installParams *InstallParameters) error {
	// Start kaltura:
	kalturaPath := filepath.Join(installParams.InstallDir, "KalturaClassroom.exe")

	cmd := exec.Command(kalturaPath)
	if err := cmd.Start(); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	// Kill it:
	if err := cmd.Process.Kill(); err != nil {
		return err
	}

	return nil
}

func main() {

	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	// err = installMSI(config.BinaryParameters, config.InstallParameters)
	// if err != nil {
	// 	log.Fatal("[ERROR] failed to install kaltura ", err)
	// }

	err = generateKalturaConfig(config.InstallParameters)
	if err != nil {
		log.Fatal("failed to generate kaltura config", err)
	}

	// Kaltura config path
	localSettingsPath := filepath.Join(os.Getenv("SystemDrive"), "\\Program Files\\Kaltura\\Classroom\\Settings\\localSettings.json")
	if err != nil {
		log.Fatal(err)
	}

	serialNumber, err := exec.Command("powershell.exe", "gwmi win32_bios serialnumber | Select -ExpandProperty serialnumber").Output()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(string(localSettingsPath)); err != nil {
		log.Fatal("unable to find kaltura local settings (localSettings.json)")
	}

	// Grab kaltura settings
	kaltura := getKalturaConfig(string(localSettingsPath))
	resourceID := int(((kaltura["config"].(map[string]interface{}))["shared"].(map[string]interface{}))["resourceId"].(float64))

	if _, err := os.Stat("credentials.json"); err != nil {
		log.Fatal("missing Google API credentials (credentials.json)")
	}

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	gConfig, err := google.ConfigFromJSON(b, config.SheetConfig.Scopes)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v", err)
	}
	client := getClient(gConfig)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("unable to retrieve Sheets client: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(config.SheetConfig.SpreadsheetID, config.SheetConfig.SheetRange).Do()
	if err != nil {
		log.Fatalf("unable to retrieve data from sheet: %v", err)
	}

	// Serial number is row[7]
	// ResourceID is row[0]
	// There should be an automated system to get these from the headers at some point

	// Make sure there is data
	if len(resp.Values) == 0 {
		log.Fatal("no data found.")
	} else {
		for _, row := range resp.Values {
			if row[7].(string) == string(serialNumber) {
				temp, _ := strconv.Atoi(row[0].(string))
				if temp != resourceID && temp == 0 {
					row[1] = resourceID

					_, err := srv.Spreadsheets.Values.Update(
						config.SheetConfig.SpreadsheetID,
						config.SheetConfig.SheetRange,
						resp,
					).ValueInputOption("USER_ENTERED").Do()

					if err != nil {
						log.Fatalf("[ERROR] unable to update sheet %v", err)
					}

					log.Println("[INFO] cells updated")
					return
				} else if intRow, _ := strconv.Atoi(row[0].(string)); intRow != resourceID {
					log.Println("[INFO] changing local settings to reflect spreadsheet")

					kaltura["config"].(map[string]interface{})["shared"].(map[string]interface{})["resourceId"], _ = strconv.Atoi(row[0].(string))

					// Update kaltura json
					updateKalturaSettings(localSettingsPath, kaltura)

					return

				} else {
					log.Println("[INFO] nothing to change for " + row[7].(string))
					return
				}
			}
		}

		// Serial Number isn't in google sheet
		// Add numbers to google sheet

		campus, err := grabRegStuff("Campus")
		if err != nil {
			log.Fatal("[ERROR] ", err)
		}

		building, err := grabRegStuff("Building")
		if err != nil {
			log.Fatal("[ERROR] ", err)
		}

		room, err := grabRegStuff("Room")
		if err != nil {
			log.Fatal("[ERROR] ", err)
		}

		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal("[ERROR] ", err)
		}

		ip, err := externalIP() 
		if err != nil {
			log.Fatal("[ERROR] ", err)
		}

		mac, err := macUint64()
		if err != nil {
			log.Fatal("[ERROR] ", err)
		}

		rb := &sheets.ValueRange{
			Values: [][]interface{}{
				{
					resourceID, // Resource ID
					campus, // Campus
					building, // Building
					room, // Room
					hostname, // Hostname
					ip, // IP Address
					mac, // Mac Address
					strings.TrimSpace(string(serialNumber)), // Serial Number
					nil, // Domain
					nil, // MBU
					nil, // SBU
					nil, // TBU
					nil, // Primary Contact
					nil, // Secondary Contact
					nil, // Backup Contact
				},
			},
		}

		r, err := srv.Spreadsheets.Values.Append(
			config.SheetConfig.SpreadsheetID,
			config.SheetConfig.SheetRange,
			rb,
		).ValueInputOption("USER_ENTERED").Do()

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("[INFO] serial Number (%s) added to the googlesheet\n", serialNumber)
		log.Println(r)

	}
}
