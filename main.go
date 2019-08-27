package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// Config : Global Configuration
type Config struct {
	BinaryParameters  *BinaryParameters  `json:"binary_parameters"`
	InstallParameters *InstallParameters `json:"install_parameters"`
	SheetConfig       *SheetConfig       `json:"google_sheet_config"`
	KalturaSettings   *KalturaSettings   `json:"kaltura_classroomn_localsettings"`
}

// BinaryParameters : Parameters to download kaltura binary
type BinaryParameters struct {
	URL          string `json:"url"`
	Checksum     string `json:"checksum"`
	FileLocation string `json:"file_location"`
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

// SheetConfig : Configuration for google sheet
type SheetConfig struct {
	Env          string `json:"env"`
	SpeadsheetID string `json:"speadsheet_id"`
	Scopes       string `json:"scopes"`
	SheetRange   string `json:"range"`
}

// KalturaSettings : Kaltura Classroom local settings
type KalturaSettings struct {
	ResourceID   string `json:"resourceID"`
	LaunchSilent string `json:"luanch_silent"`
	Countdown    string `json:"countdown"`
}

// Grabs the kaltura configuration file
func getKalturaConfig(path string) map[string]interface{}, error {

	var kaltura map[string]interface{}

	// Open our jsonFile
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	log.Println("[INFO] successfully opened localSettings")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// json file's content into 'config' which we defined above
	json.Unmarshal(byteValue, &kaltura)

	return kaltura, nil
}

func updateKalturaSettings(path string, newSettings map[string]interface{}) error {

	marshalledSettings, _ := json.MarshalIndent(newSettings, "", "\t")
	err := ioutil.WriteFile(path, marshalledSettings, 0644)

	return err
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

func installMSI(binParams *BinaryParameters, installParams *InstallParameters) error {

	// TODO: Download Binary

	tmplString := `/i %s /qn /norestart
		INSTALLDIR=%s
		ADDLOCAL=ALL
		KALTURA_RECORDINGS_DIR=%s
		KALTURA_URL=%s
		KALTURA_APPTOKEN=%s
		KALTURA_APPTOKEN_ID=%s
		KALTURA_PARTNER_ID=%s
		INSTALLDESKTOPSHORTCUT=%s
		INSTALLPROGRAMSSHORTCUT=%s
	`
	installString := fmt.Sprintf(tmplString,
		binParams.FileLocation,
		installParams.InstallDir,
		installParams.RecordingDir,
		installParams.URL,
		installParams.AppToken,
		installParams.AppTokenID,
		installParams.DesktopShortcut,
		installParams.ProgramShortcut,
	)

	cmd := exec.Command("msiexec.exe", installString)
	if err := cmd.Run(); err != nil {
		fmt.Println("[ERROR] could not install kaltura")
		return err
	}

	return nil
}

// func createLocalSettings(binParams BinaryParameters, installParams InstallParameters) error

func main() {

	// Find the Kaltura local settings
	localSettingsPath := filepath.Join(os.Getenv("SystemDrive"), "\\VCU-Deploy\\config\\Kaltura\\localSettings.json")

	serialNumber, err := exec.Command("powershell.exe", "gwmi win32_bios serialnumber | Select -ExpandProperty serialnumber").Output()
	if err != nil {
		log.Fatal(err)
	}

	// Make sure Google credentials file exists
	if _, err := os.Stat("credentials.json"); err != nil {
		log.Fatal("[ERROR] missing Google API credentials (credentials.json)")
	}

	// Make sure localSettings config file exists
	if _, err := os.Stat(localSettingsPath); err != nil {
		log.Fatal("[ERROR] unable to find kaltura local settings (localSettings.json)")
	}

	// Serialize JSON config file
	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Download and Install Kaltura Classroom MSI
	err = installMSI(config.BinaryParameters, config.InstallParameters)
	if err != nil {
		log.Fatal(err)
	}

	// Grab kaltura settings
	kaltura, err := getKalturaConfig(string(localSettingsPath))
	if err != nil {
		log.Fatal(err)
	}
	
	resourceID := int(((kaltura["config"].(map[string]interface{}))["shared"].(map[string]interface{}))["resourceId"].(float64))

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("unable to read client secret file: %v\n", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	gConfig, err := google.ConfigFromJSON(b, config.SheetConfig.Scopes)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v\n", err)
	}

	client := getClient(gConfig)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("unable to retrieve Sheets client: %v\n", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(config.SheetConfig.SpeadsheetID, config.SheetConfig.SheetRange).Do()
	if err != nil {
		log.Fatalf("unable to retrieve data from sheet: %v\n", err)
	}

	// Make sure there is data
	if len(resp.Values) == 0 {
		log.Fatal("[ERROR] No data found.")
	} else {
		for i, row := range resp.Values {

			// Serial Number is already in google sheet
			if row[0].(string) == string(serialNumber) {
				temp, _ := strconv.Atoi(row[20].(string))
				if temp != resourceID && temp == 0 {
					row[20] = resourceID

					result, err := srv.Spreadsheets.Values.Update(
						config.SheetConfig.SpeadsheetID,
						config.SheetConfig.SheetRange,
						resp,
					).ValueInputOption("USER_ENTERED").Do()

					if err != nil {
						log.Fatalf("unable to update sheet %v", err)
					}

					log.Printf("[INFO] %d cells updated\n", result.UpdatedCells)
					return
				} else if intRow, _ := strconv.Atoi(row[20].(string)); intRow != resourceID {
					log.Println("[INFO] changing local settings to reflect spreadsheet")

					kaltura["config"].(map[string]interface{})["shared"].(map[string]interface{})["resourceId"], err = strconv.Atoi(row[20].(string))
					if err != nil {
						log.Fatal(err)
					}

					err = updateKalturaSettings(localSettingsPath, kaltura)
					if err != nil {
						log.Fatal(err)
					}

					return

				} else {
					log.Printf("[INFO] %d.) nothing to change for %s\n", i, row[0].(string))

					return
				}
			}
		}

		// Serial Number isn't in google sheet
		// Add Serial Number to google sheet

		rb := &sheets.ValueRange{
			Values: [][]string{
				{
					serialNumber, 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"", 
					"",
					resourceID,
				},
			}
		}

		resp, err = srv.SpreadSheet.Values.Append(
			config.SheetConfig.SpeadsheetID,
			config.SheetConfig.SheetRange,
			rb,
		).ValueInputOption("USER_ENTERED").Do()

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Serial Number (%s) added to the googlesheet\n", serialNumber)
		fmt.Println(resp)

	}
}
