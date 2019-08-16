package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"lo"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("[ERROR] unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("[ERROR] unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
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

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// json file's content into 'config' which we defined above
	json.Unmarshal(byteValue, &kaltura)

	return kaltura
}

func updateKalturaSettings(path []byte, newSettings map[string]interface{}) {

	marshalledSettings, _ := json.MarshalIndent(newSettings, "", "\t")
	err := ioutil.WriteFile(string(path), marshalledSettings, 0644)

	if err != nil {
		log.Println("[ERROR] new kaltura config couldn't be written to")
	}
}

func getConfig() *Config {

	var config Config

	// Open our jsonFile
	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[INFO] successfully opened config.json")

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'config' which we defined above
	json.Unmarshal(byteValue, &config)

	return &config
}

func installMSI(binParams BinaryParameters, installParams InstallParameters) error {

	tmplString := `msiexec.exe /i %s /qn /norestart
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
	installString := fmt.Sprintf()
		
	return errors.New("Placeholder")
}

func createLocalSettings(binParams BinaryParameters, installParams InstallPararmeters) error {
	
}

func main() {

	var serialNumber []byte
	var localSettingsPath []byte

	config := getConfig()

	if config.SheetConfig.Env == "dev" {
		serialNumber = []byte("3WFZBH2")
		localSettingsPath = []byte("localSettings.json")
	} else if config.SheetConfig.Env == "prod" {

		// Find the Kaltura local settings
		houstonsConfigPath := filepath.Join(os.Getenv("SystemDrive"), "\\VCU-Deploy\\config\\Kaltura\\config.ps1")

		var err error
		localSettingsPath, err = exec.Command("powershell.exe", houstonsConfigPath).Output()
		if err != nil {
			log.Fatal(err)
		}

		serialNumber, err = exec.Command("powershell.exe", "gwmi win32_bios serialnumber | Select -ExpandProperty serialnumber").Output()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println(config.SheetConfig.Env)
		log.Fatal("[ERROR] unknown 'Env' in configuration file (must be 'dev' or 'prod') or environment variables not set properly")
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

	resp, err := srv.Spreadsheets.Values.Get(config.SheetConfig.SpeadsheetID, config.SheetConfig.SheetRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// Make sure there is data
	if len(resp.Values) == 0 {
		log.Fatal("[ERROR] No data found.")
	} else {
		for _, row := range resp.Values {

			// Serial Number is already in google sheet
			if row[0].(string) == string(serialNumber) {
				temp, _ := strconv.Atoi(row[20].(string))
				if temp != resourceID && temp == 0 {
					row[20] = resourceID

					_, err := srv.Spreadsheets.Values.Update(
						config.SheetConfig.SpeadsheetID,
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

	}
}
