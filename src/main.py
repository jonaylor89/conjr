#!/usr/bin/env python3

import sys
import json
import pickle
import os

from subprocess import check_output

from googleapiclient.discovery import build
from google_auth_oauthlib.flow import InstalledAppFlow
from google.auth.transport.requests import Request


def main():
    """
    Shows basic usage of the Sheets API.
    Prints values from a sample spreadsheet.
    """

    # Check for configuration file
    if not os.path.exists("config.json"):
        print("[ERROR] missing configuration file (config.json)")
        return

    # Load configurations
    with open("config.json") as f:
        config = json.load(f)

    scopes = config["scopes"]
    spreadsheet_id = config["speadsheet_id"]
    range = config["range"]

    serial_number = None

    if config["env"] == "dev":
        # Hardcoded for now
        serial_number = "3WFZBH2" 
    elif config["env"] == "prod":
        # The powershell to grab the serial number in production
        serial_number = check_output(["powershell.exe", "gwmi win32_bios serialnumber | Select -ExpandProperty serialnumber"])
    else:
        print("[ERROR] Unknown 'env' in configuration file (must be 'dev' or 'prod')")
        return

    if config["env"] == "dev":
        local_settings_path = "localSettings.json"
    elif config["env"] == "prod":
        # Find the Kaltura local settings
        houstins_config_path = os.path.join(os.getenv("SystemDrive"), "\\VCU-Deploy\\config\\Kaltura\\config.ps1")
    
        local_settings_path = check_output(["powershell.exe", houstins_config_path])
    else:
        print("[ERROR] 'env' in configuration file (must be 'dev' or 'prod')")
        return

    if not os.path.exists(local_settings_path):
        print("[ERROR] unable to find kaltura local settings (localSettings.json)")
        return

    with open(local_settings_path) as f:
        kaltura = json.load(f)

    resource_id = str(kaltura["config"]["shared"]["resourceId"])

    # Check for google API credentials
    if not os.path.exists("credentials.json"):
        print("[ERROR] missing Google API credentials (credentials.json)")
        return

    creds = None
    # The file token.pickle stores the user's access and refresh tokens, and is
    # created automatically when the authorization flow completes for the first
    # time.
    if os.path.exists("token.pickle"):
        with open("token.pickle", "rb") as token:
            creds = pickle.load(token)

    # If there are no (valid) credentials available, let the user log in.
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            flow = InstalledAppFlow.from_client_secrets_file("credentials.json", scopes)
            creds = flow.run_local_server()

        # Save the credentials for the next run
        with open("token.pickle", "wb") as token:
            pickle.dump(creds, token)

    service = build("sheets", "v4", credentials=creds)

    # Call the Sheets API
    sheet = service.spreadsheets()
    result = sheet.values().get(spreadsheetId=spreadsheet_id, range=range).execute()
    values = result.get("values", [])

    # Make sure there is data
    if values:
        # Exclude the first row because it is full of headers
        for row in values[1:]:
            if row[0] == serial_number:

                # 19 is where the rID is and 0 is the default value meaning nothing has been set
                if row[19] != resource_id and int(row[19]) == 0:

                    # Update rID
                    row[19] = resource_id

                    body = {"values": values}
                    result = (
                        service.spreadsheets()
                        .values()
                        .update(
                            spreadsheetId=spreadsheet_id,
                            range=range,
                            valueInputOption="USER_ENTERED",
                            body=body,
                        )
                        .execute()
                    )

                    print("[INFO] cells updated.")
                    return

                elif row[19] != resource_id:
                    print("[INFO] changing local settings to reflect spreadsheet")

                    kaltura["config"]["shared"]["resourceId"] = int(row[19])

                    # Reopen localSettings to update resourceId
                    with open("localSettings.json", "w") as f:
                        json.dump(kaltura, f, indent=2)

                    return 

                else:
                    print(f"[INFO] nothing to change for {row[0]}")

                    return
        else:
            print("[ERROR] could not find serial number in sheet")
    else:
        print("[ERROR] no data found")


if __name__ == "__main__":
    main()
