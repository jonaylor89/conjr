#!/usr/bin/env python3

import sys
import pickle
import os.path
from googleapiclient.discovery import build
from google_auth_oauthlib.flow import InstalledAppFlow
from google.auth.transport.requests import Request

# If modifying these scopes, delete the file token.pickle.
SCOPES = ["https://www.googleapis.com/auth/spreadsheets"]

# The ID and range of a sample spreadsheet.
SAMPLE_SPREADSHEET_ID = "1WG9lDcABDF0FU84QofkgAWoilzEedt2mjibTBKHr7Rs"
RANGE = "A:T"


def main():
    """
    Shows basic usage of the Sheets API.
    Prints values from a sample spreadsheet.
    """

    if not os.path.exists("credentials.json"):
        print("[ERROR] Missing Google API credentials (credentials.json)")
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
            flow = InstalledAppFlow.from_client_secrets_file("credentials.json", SCOPES)
            creds = flow.run_local_server()

        # Save the credentials for the next run
        with open("token.pickle", "wb") as token:
            pickle.dump(creds, token)

    service = build("sheets", "v4", credentials=creds)

    # Call the Sheets API
    sheet = service.spreadsheets()
    result = (
        sheet.values().get(spreadsheetId=SAMPLE_SPREADSHEET_ID, range=RANGE).execute()
    )
    values = result.get("values", [])

    if not values:
        print("[ERROR] No data found.")
    else:
        for row in values[1:]:
            if row[0] == "3WG0CH2":

                # 19 is where the rID is
                if row[19] != "THE BEST ID":

                    # Update rID
                    row[19] = "THE BEST ID"

                    body = {"values": values}
                    result = (
                        service.spreadsheets()
                        .values()
                        .update(
                            spreadsheetId=SAMPLE_SPREADSHEET_ID,
                            range=RANGE,
                            valueInputOption="USER_ENTERED",
                            body=body,
                        )
                        .execute()
                    )

                    print("[INFO] cells updated.")
                    return 

                else:
                    print(row)
                    return


if __name__ == "__main__":
    main()
