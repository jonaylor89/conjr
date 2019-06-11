#!/usr/bin/env python3

import sys
import json
import pickle
import os.path
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
    config = json.loads("config.json")

    scopes = config["scopes"]
    spreadsheet_id = config["speadsheet_id"]
    range = config["range"]

    # Check for google API creds
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
        for row in values[1:]:
            if row[0] == "3WG0CH2":  # Value with be from kaltura json

                # 19 is where the rID is
                if row[19] != "THE BEST ID":  # Value will be from kaltura json

                    # Update rID
                    row[19] = "THE BEST ID"  # Value will be from kaltura json

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

                else:
                    print(f"[INFO] nothing to change for {row[0]}")

                    # TODO: I think Houstin wants me to edit the kaltura json file here?

                    return
    else:
        print("[ERROR] no data found")


if __name__ == "__main__":
    main()
