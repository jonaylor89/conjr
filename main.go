package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

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

	log.Println("[INFO] successfully Opened localSettings")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// json file's content into 'config' which we defined above
	json.Unmarshal(byteValue, &kaltura)

	return kaltura
}

func updateKalturaSettings(path []byte, newSettings map[string]interface{}) error {

	marshalledSettings, _ := json.MarshalIndent(newSettings, "", "\t")
	err := ioutil.WriteFile(string(path), marshalledSettings, 0644)

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

	log.Println("[INFO] Successfully Opened config.json")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'config' which we defined above
	json.Unmarshal(byteValue, &config)

	return &config, nil
}

// downloadFile will download a url to a local file
func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func installMSI(binParams *BinaryParameters, installParams *InstallParameters) error {

	// Download Binary
	err := downloadFile(binParams.FileLocation, binParams.URL)
	if err != nil {
		return err
	}

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

func main() {

	var serialNumber []byte
	var localSettingsPath []byte

	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Find the Kaltura local settings
	houstonsConfigPath := filepath.Join(os.Getenv("SystemDrive"), "\\VCU-Deploy\\config\\Kaltura\\config.ps1")

	localSettingsPath, err = exec.Command("powershell.exe", houstonsConfigPath).Output()
	if err != nil {
		log.Fatal(err)
	}

	serialNumber, err = exec.Command("powershell.exe", "gwmi win32_bios serialnumber | Select -ExpandProperty serialnumber").Output()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(string(localSettingsPath)); err != nil {
		log.Fatal("[ERROR] unable to find kaltura local settings (localSettings.json)")
	}

	// Grab kaltura settings
	kaltura := getKalturaConfig(string(localSettingsPath))
	resourceID := int(((kaltura["config"].(map[string]interface{}))["shared"].(map[string]interface{}))["resourceId"].(float64))

	if _, err := os.Stat("credentials.json"); err != nil {
		log.Fatal("[ERROR] missing Google API credentials (credentials.json)")
	}

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	gConfig, err := google.ConfigFromJSON(b, config.SheetConfig.Scopes)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(gConfig)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(config.SheetConfig.SpreadsheetID, config.SheetConfig.SheetRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// Make sure there is data
	if len(resp.Values) == 0 {
		log.Fatal("[ERROR] No data found.")
	} else {
		for _, row := range resp.Values {
			if row[0].(string) == string(serialNumber) {
				temp, _ := strconv.Atoi(row[20].(string))
				if temp != resourceID && temp == 0 {
					row[20] = resourceID

					_, err := srv.Spreadsheets.Values.Update(
						config.SheetConfig.SpreadsheetID,
						config.SheetConfig.SheetRange,
						resp,
					).ValueInputOption("USER_ENTERED").Do()

					if err != nil {
						log.Fatalf("unable to update sheet %v", err)
					}

					log.Println("[INFO] cells updated")
					return
				} else if intRow, _ := strconv.Atoi(row[20].(string)); intRow != resourceID {
					log.Println("[INFO] changing local settings to reflect spreadsheet")

					kaltura["config"].(map[string]interface{})["shared"].(map[string]interface{})["resourceId"], _ = strconv.Atoi(row[20].(string))

					// TODO: Update kaltura json
					updateKalturaSettings(localSettingsPath, kaltura)

				} else {
					log.Println("[INFO] nothing to change for " + row[0].(string))
					return
				}
			}
		}

		// Serial Number isn't in google sheet
		// Add Serial Number to google sheet

		rb := &sheets.ValueRange{
			Values: [][]interface{}{
				{
					serialNumber,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					resourceID,
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

		fmt.Printf("Serial Number (%s) added to the googlesheet\n", serialNumber)
		fmt.Println(r)

	}
}
