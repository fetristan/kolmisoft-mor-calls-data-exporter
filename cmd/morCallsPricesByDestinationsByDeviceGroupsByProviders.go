package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
	"github.com/pariz/gountries"
	"github.com/spf13/cobra"
)

// Initialize the command.
func init() {
	rootCmd.AddCommand(morCallsPricesByDestinationsByDeviceGroupsByProvidersCmd)
	morCallsPricesByDestinationsByDeviceGroupsByProvidersCmd.Flags().StringP("dateStart", "s", "", "The start date of the export")
	morCallsPricesByDestinationsByDeviceGroupsByProvidersCmd.Flags().StringP("dateEnd", "e", "", "The end date of the export")
}

// Model for MOR call prices by destinations, device groups, and providers.
type ModelMorCallsPricesByDestinationsByDeviceGroupsByProviders struct {
	DeviceGroup string
	Destination string
	Prefix      string
	Price       string
	Duration    int
}

// Retrieve call data from the database and return it as a slice of models.
func getModelMorCallsPricesByDestinationsByDeviceGroupsByProviders(stmt *sql.Stmt) ([]any, error) {
	var messages []any

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var msg ModelMorCallsPricesByDestinationsByDeviceGroupsByProviders

		if err := rows.Scan(
			&msg.DeviceGroup,
			&msg.Destination,
			&msg.Prefix,
			&msg.Price,
			&msg.Duration,
		); err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// Define the main Cobra command for exporting call prices.
var morCallsPricesByDestinationsByDeviceGroupsByProvidersCmd = &cobra.Command{
	Use:   "morCallsPricesByDestinationsByDeviceGroupsByProviders",
	Short: "Export the prices of the calls from MOR database by destination grouped by device groups filtered by providers and devices.",
	Long: `Export call prices from the MOR database, grouped by device groups, filtered by providers and devices, and organized by destination. The CSV will include the following columns: Device Group, Country, Destination, Prefix, Price, Duration, Duration (hours), Average (Price/Min).

Usage:
  morCallsPricesByDestinationsByDeviceGroupsByProviders -s [start_date] -e [end_date]

Flags:
  -s, --dateStart string   The start date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')
  -e, --dateEnd string     The end date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')

Example:
  morCallsPricesByDestinationsByDeviceGroupsByProviders -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"

This command exports call prices from the MOR database, grouped by device groups, filtered by providers and devices, and organized by destination. It generates a CSV file with the specified start and end date. The CSV file contains information about Device Group, Country, Destination, Prefix, Price, Duration, Duration (in hours), and Average Price per Minute. The generated CSV file is named with a timestamp and saved in the current working directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Obtain the start and end date strings from the command-line flags.
		dateStartStr, _ := cmd.Flags().GetString("dateStart")
		dateEndStr, _ := cmd.Flags().GetString("dateEnd")

		// Parse the provided start and end dates.
		dateStart, err := time.Parse("2006-01-02 15:04:05", dateStartStr)
		if err != nil {
			fmt.Println("Invalid dateStart format. Please use 'YYYY-MM-DD HH:mm:SS'")
			return
		}

		dateEnd, err := time.Parse("2006-01-02 15:04:05", dateEndStr)
		if err != nil {
			fmt.Println("Invalid dateEnd format. Please use 'YYYY-MM-DD HH:mm:SS'")
			return
		}

		// Display the start and end date information for the user's reference.
		fmt.Println("morCallsPricesByDestinationsByDeviceGroupsByProviders called with dateStart: " + dateStart.Format("2006-01-02 15:04:05") + " and dateEnd: " + dateEnd.Format("2006-01-02 15:04:05"))

		// Define mapping of device groups to source devices.
		srcGroupDevicesID := map[string][]string{
			"EN": {"181", "1081"},
			"FR": {"671", "1072"},
		}

		// List of provider IDs.
		providersID := []string{"561", "721", "21", "31", "101", "111", "441", "711", "781", "801"}

		// Initialize variables to store SQL filters and device IDs.
		srcDevicesIDFilter := ""
		srcDevicesIDList := ""

		// Build SQL filters based on the mapping of device groups to source devices.
		for oneSrcGroup, oneSrcGroupDevicesID := range srcGroupDevicesID {
			srcDevicesIDFilter += fmt.Sprintf("			WHEN src_device_id IN (%s) THEN '%s' \n", strings.Join(oneSrcGroupDevicesID, ","), oneSrcGroup)
			srcDevicesIDList += strings.Join(oneSrcGroupDevicesID, ",")
		}

		// Construct the SQL query with placeholders.
		request := fmt.Sprintf(`
		SELECT
		CASE
%s
        	END AS DeviceGroup,
        	mor.destinations.name AS Destination,
        	c.prefix as Prefix,
        	REPLACE(CAST(ROUND(SUM(provider_price), 2) AS CHAR), '.', ',') AS Price,
		ROUND(SUM(duration)/60) AS Duration 
		FROM mor.calls c inner join mor.destinations on c.prefix = mor.destinations.prefix
		WHERE 
			calldate  > '%s' AND
			calldate  < '%s' AND
			src_device_id IN (%s) AND
			provider_id IN (%s)
		GROUP BY DeviceGroup, destination
		ORDER BY DeviceGroup, destination;`, srcDevicesIDFilter, dateStartStr, dateEndStr, srcDevicesIDList, strings.Join(providersID, ","))

		// Log the SQL query for debugging and tracking purposes.
		log.Print(request)

		// Send the SQL request to the MorRequest function and obtain results.
		results, err := MorRequest(request, getModelMorCallsPricesByDestinationsByDeviceGroupsByProviders)
		if err != nil {
			log.Fatal(err)
		}

		// Create a list to hold the results casted to the desired data model.
		var resultsCasted []ModelMorCallsPricesByDestinationsByDeviceGroupsByProviders
		for _, result := range results {
			resultsCasted = append(resultsCasted, result.(ModelMorCallsPricesByDestinationsByDeviceGroupsByProviders))
		}

		// Generate a filename for the output file.
		now := time.Now()
		filename := fmt.Sprintf("%d_%02d_%02d_%02d_%02d_%02d_export.csv", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

		// Create and open the output file for writing.
		outputFile, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFile.Close()

		// Write the header row to the output file.
		fmt.Fprintln(outputFile, "Device group;Country;Destination;Prefix;Price;Duration;Duration (hours);Average (Price/Min)")

		// Process and write each result to the output file.
		for _, oneResult := range resultsCasted {
			// Format the duration to hours, minutes, and seconds.
			durationHourMinSeconds := formatTimeMinutesToHours(oneResult.Duration)

			// Calculate the average price per minute.
			averagePriceMinFloat := float32(0.0)

			// Handle potential decimal comma.
			priceComma := strings.Replace(oneResult.Price, ",", ".", 1)
			priceFloat, err := strconv.ParseFloat(priceComma, 64)
			if err != nil {
				log.Fatal(err)
			}

			if priceFloat > 0 && oneResult.Duration > 0 {
				averagePriceMinFloat = float32(priceFloat / float64(oneResult.Duration))
			}

			// Format the average price per minute.
			averagePriceMin := strconv.FormatFloat(float64(averagePriceMinFloat), 'f', -1, 32)

			// Truncate the average price if it has more than 4 decimal places.
			if len(averagePriceMin) > 4 {
				averagePriceMin = averagePriceMin[:4]
			}

			// Process and format the prefix for phone numbers.
			oneResult.Prefix = "+" + oneResult.Prefix
			oneResult.Prefix = strings.TrimRight(oneResult.Prefix+"000000", " ")[0:6]

			// Parse and format phone number information.
			phoneNumber, err := phonenumbers.Parse(oneResult.Prefix, "")
			if err == nil {
				// Format the phone number in international format using the phonenumbers library.
				formattedPrefix := phonenumbers.Format(phoneNumber, phonenumbers.INTERNATIONAL)

				// Retrieve the region (country) code associated with the phone number.
				regionCode := phonenumbers.GetRegionCodeForNumber(phoneNumber)

				// Initialize a new query object for the gountries library, which will be used to fetch additional country information.
				query := gountries.New()

				// Find the country information based on the region code.
				displayRegionQuery, _ := query.FindCountryByAlpha(regionCode)
				displayRegion := displayRegionQuery.Name.Common

				if displayRegion == "" {
					// If the region is not found, provide fallback information for certain countries.
					if strings.Contains(strings.ToLower(oneResult.Destination), "australia") || strings.Contains(strings.ToLower(oneResult.Destination), "australie") {
						displayRegion = "Australia"
					} else if strings.Contains(strings.ToLower(oneResult.Destination), "canada") {
						displayRegion = "Canada"
					} else if strings.Contains(strings.ToLower(oneResult.Destination), "italy") {
						displayRegion = "Italy"
					} else if strings.Contains(strings.ToLower(oneResult.Destination), "russia") {
						displayRegion = "Russia"
					} else if strings.Contains(strings.ToLower(oneResult.Destination), "unites states") {
						displayRegion = "United States"
					} else if strings.Contains(strings.ToLower(oneResult.Destination), "guadeloupe") {
						displayRegion = "France"
					} else if strings.Contains(strings.ToLower(oneResult.Destination), "morocco") {
						displayRegion = "Morocco"
					} else if strings.Contains(strings.ToLower(oneResult.Destination), "reunion") || strings.Contains(strings.ToLower(oneResult.Destination), "r?union") || strings.Contains(strings.ToLower(oneResult.Destination), "france") {
						displayRegion = "France"
					} else if strings.Contains(strings.ToLower(oneResult.Destination), "uk") || strings.Contains(strings.ToLower(oneResult.Destination), "united kingdom") {
						displayRegion = "United Kingdom"
					}
				}

				if displayRegion == "" {
					displayRegion = "UNKNOWN"
				}

				// Write the formatted result to the output file.
				fmt.Fprintf(outputFile, "%s;%s;%s;%s;%s;%d;%s;%s\n", oneResult.DeviceGroup, displayRegion, oneResult.Destination, formattedPrefix, oneResult.Price, oneResult.Duration, durationHourMinSeconds, averagePriceMin)
			} else {
				// Handle the case where phone number information cannot be parsed.
				displayRegion := "UNKNOWN"
				formattedPrefix := "UNKNOWN"

				// Write the result with unknown information to the output file.
				fmt.Fprintf(outputFile, "%s;%s;%s;%s;%s;%d;%s;%s\n", oneResult.DeviceGroup, displayRegion, oneResult.Destination, formattedPrefix, oneResult.Price, oneResult.Duration, durationHourMinSeconds, averagePriceMin)
			}
		}
		// Log a message indicating the filename of the exported data.
		log.Printf("%s exported", filename)
	},
}
