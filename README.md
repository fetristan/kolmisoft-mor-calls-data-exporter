# Kolmisoft MOR calls data exporter

This command-line tool to exports data from Kolmisoft Mor.

## Installation

Before using this tool, make sure you have Go installed on your system.

To install the tool, run the following command:

```bash
git clone https://github.com/fetristan/kolmisoft-mor-calls-data-exporter.git
cd kolmisoft-mor-calls-data-exporter
go mod tidy
go run main.go
```

Or just download the binary file and execute it.

## Configuration
Create a .env file at the root directory of the project and add the following variables (update with your credentials and settings):

    DB_IP_MOR=127.0.0.1
    DB_IP_MOR=127.0.0.1
    DB_PORT_MOR=3306
    DB_NAME_MOR=mor
    DB_USER_MOR=mor
    DB_PASS_MOR=mor
    DB_SSH_IP_MOR=192.168.0.2
    DB_SSH_PORT_MOR=22
    DB_SSH_USER_MOR=user
    DB_SSH_KEY_MOR=/home/user/.ssh/id_rsa
    DB_SSH_KEY_PASS_MOR=ThePassword

## Usage

# morCallsPricesByDestinationsByDeviceGroupsByProviders usage:

```bash
go run main.go morCallsPricesByDestinationsByDeviceGroupsByProviders -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"
or execute the binary file and morCallsPricesByDestinationsByDeviceGroupsByProviders -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"
```

This command exports call prices from the MOR database, grouped by device groups, filtered by providers and devices, and organized by destination. The generated CSV file is named with a timestamp and saved in the current working directory.

In the file morCallsPricesByDestinationsByDeviceGroupsByProviders.go, change the following variables to match your MOR database device id / provider id and to choose the name of your device group:

    srcGroupDevicesID
    providersID

You can use the morCallsPricesByDestinationsByDeviceGroupsByProviders command with the following options:
```bash
    -s, --dateStart (string): The start date of the export (e.g., 'YYYY-MM-DD HH:mm:SS').
    -e, --dateEnd (string): The end date of the export (e.g., 'YYYY-MM-DD HH:mm:SS').
```

The exported CSV file contains the following columns:

    Pole
    Country
    Destination
    Prefix
    Price
    Duration
    Duration (hours)
    Average (Price/Min)

# morIncomingCallsDuration usage:

```bash
go run main.go morIncomingCallsDuration -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"
or execute the binary file and morIncomingCallsDuration -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"
```

This command Export incoming call duration data for a specified date range. The generated CSV file is named with a timestamp and saved in the current working directory.

You can use the morIncomingCallsDuration command with the following options:
```bash
    -s, --dateStart (string): The start date of the export (e.g., 'YYYY-MM-DD HH:mm:SS').
    -e, --dateEnd (string): The end date of the export (e.g., 'YYYY-MM-DD HH:mm:SS').
```

The exported CSV file contains the following columns:

    Did
    Seconds
    Provider
    Username
    Extension
    Description
    Status
    UpdateDate
    Duration (hours)

# morCallsDurationPerMobileOrLandlinePhones usage:

```bash
go run main.go morCallsDurationPerMobileOrLandlinePhones -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"
or execute the binary file and morCallsDurationPerMobileOrLandlinePhones -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"
```

This command export calls duration per mobile or landline phones for a specified date range. The generated CSV file is named with a timestamp and saved in the current working directory.

You can use the morCallsDurationPerMobileOrLandlinePhones command with the following options:
```bash
    -s, --dateStart (string): The start date of the export (e.g., 'YYYY-MM-DD HH:mm:SS').
    -e, --dateEnd (string): The end date of the export (e.g., 'YYYY-MM-DD HH:mm:SS').
```

The exported CSV file contains the following columns:

    Country
    Destination
    Duration
    Duration (hours)

## Acknowledgements

This tool uses the following libraries:

    nyaruka/phonenumbers for phone number parsing and formatting.
    pariz/gountries for additional country information.
    spf13/cobra for the command-line interface.
    spf13/viper for configuration management.

## License

This tool is open-source and available under the MIT License.
