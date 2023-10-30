package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nyaruka/phonenumbers"
	"github.com/spf13/cobra"
)

// Initialize the command.
func init() {
	rootCmd.AddCommand(morCallsDurationPerMobileOrLandlinePhones)
	morCallsDurationPerMobileOrLandlinePhones.Flags().StringP("dateStart", "s", "", "The start date of the export")
	morCallsDurationPerMobileOrLandlinePhones.Flags().StringP("dateEnd", "e", "", "The end date of the export")
}

// ModelMorCallsDurationPerMobileOrLandlinePhones represents information about incoming calls duration.
type ModelMorCallsDurationPerMobileOrLandlinePhones struct {
	Destination string
	Duration    int
}

// Retrieve call data from the database and return it as a slice of models.
func getModelMorCallsDurationPerMobileOrLandlinePhones(stmt *sql.Stmt) ([]any, error) {
	var messages []any

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var msg ModelMorCallsDurationPerMobileOrLandlinePhones

		if err := rows.Scan(
			&msg.Destination,
			&msg.Duration,
		); err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// Define the main Cobra command for exporting call prices.
var morCallsDurationPerMobileOrLandlinePhones = &cobra.Command{
	Use:   "morCallsDurationPerMobileOrLandlinePhones",
	Short: "Export answered outgoing calls duration per mobile or landline phones for a specified date range.",
	Long: `Export answered outgoing calls duration per mobile or landline phones for a specified date range. The CSV will include the following columns: Country, Destination, Duration, Duration (hours).

Usage:
  morCallsDurationPerMobileOrLandlinePhones -s [start_date] -e [end_date]

Flags:
  -s, --dateStart string   The start date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')
  -e, --dateEnd string     The end date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')

Example:
  morCallsDurationPerMobileOrLandlinePhones -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"

This command export answered outgoing calls duration per mobile or landline phones for a specified date range. It generates a CSV file with the specified start and end date. The CSV will include the following columns: Country, Destination, Duration, Duration (hours). The generated CSV file is named with a timestamp and saved in the current working directory.`,
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
		fmt.Println("morCallsDurationPerMobileOrLandlinePhones called with dateStart: " + dateStart.Format("2006-01-02 15:04:05") + " and dateEnd: " + dateEnd.Format("2006-01-02 15:04:05"))

		// Construct the SQL query with placeholders.
		request := fmt.Sprintf(`SELECT
		dst as Destination,
		billsec as Duration
		FROM mor.calls
		WHERE calldate > '%s' and calldate < '%s' and dst_device_id = 0 and disposition = 'ANSWERED';`, dateStartStr, dateEndStr)

		// Log the SQL query for debugging and tracking purposes.
		log.Print(request)

		// Send the SQL request to the MorRequest function and obtain results.
		results, err := MorRequest(request, getModelMorCallsDurationPerMobileOrLandlinePhones)
		if err != nil {
			log.Fatal(err)
		}

		// Create a list to hold the results casted to the desired data model.
		var resultsCasted []ModelMorCallsDurationPerMobileOrLandlinePhones
		for _, result := range results {
			resultsCasted = append(resultsCasted, result.(ModelMorCallsDurationPerMobileOrLandlinePhones))
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
		fmt.Fprintln(outputFile, "Country;Destination;Duration;Duration (hours)")

		// Process and write each result to the output file.
		for _, oneResult := range resultsCasted {
			// Format the duration to hours, minutes, and seconds.
			durationHourMinSeconds := formatTimeSecondsToHours(oneResult.Duration)

			// Add a leading plus sign to the destination number.
			oneResult.Destination = "+" + oneResult.Destination

			// Remove the leading zeros from the destination number.
			oneResult.Destination = removeZero(oneResult.Destination)

			// Parse and format phone number information.
			phoneNumber, err := phonenumbers.Parse(oneResult.Destination, "")
			if err == nil {

				// Retrieve the region (country) code associated with the phone number.
				regionCode := phonenumbers.GetRegionCodeForNumber(phoneNumber)

				// Retrieve the type of the number.
				phoneNumberType := phonenumbers.GetNumberType(phoneNumber)

				// Append the phone number type to the region code if it's mobile.
				if phoneNumberType == phonenumbers.MOBILE {
					regionCode += "_MOBILE"
				}

				// Write the formatted result to the output file.
				fmt.Fprintf(outputFile, "%s;%s;%d;%s\n", regionCode, oneResult.Destination, oneResult.Duration, durationHourMinSeconds)
			}

		}
		// Log a message indicating the filename of the exported data.
		log.Printf("%s exported", filename)
	},
}
