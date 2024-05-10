package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
	"github.com/pariz/gountries"
	"github.com/spf13/cobra"
)

// Initialize the command.
func init() {
	rootCmd.AddCommand(morMaxCallsNumberPerDaysByDestinations)
	morMaxCallsNumberPerDaysByDestinations.Flags().StringP("dateStart", "s", "", "The start date of the export")
	morMaxCallsNumberPerDaysByDestinations.Flags().StringP("dateEnd", "e", "", "The end date of the export")
}

// Model for MOR call prices by destinations, device groups, and providers.
type ModelMorMaxCallsNumberPerDaysByDestinations struct {
	Day         string
	Destination string
	Prefix      string
	Calls       int
}

// Model for MOR call prices by country, device groups, and providers.
type ModelMorMaxCallsNumberPerDaysByCountry struct {
	Day     string
	Country string
	Calls   int
}

// Retrieve call data from the database and return it as a slice of models.
func getModelMorMaxCallsNumberPerDaysByDestinations(stmt *sql.Stmt) ([]any, error) {
	var messages []any

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var msg ModelMorMaxCallsNumberPerDaysByDestinations

		if err := rows.Scan(
			&msg.Day,
			&msg.Destination,
			&msg.Prefix,
			&msg.Calls,
		); err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// Define the main Cobra command for exporting call prices.
var morMaxCallsNumberPerDaysByDestinations = &cobra.Command{
	Use:   "morMaxCallsNumberPerDaysByDestinations",
	Short: "Export the maximum numbers of calls for each destinations per days.",
	Long: `Export the maximum numbers of calls for each destinations per days from the MOR database, the CSV will include the following columns: Day, Country, Calls.

Usage:
morMaxCallsNumberPerDaysByDestinations -s [start_date] -e [end_date]

Flags:
  -s, --dateStart string   The start date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')
  -e, --dateEnd string     The end date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')

Example:
morMaxCallsNumberPerDaysByDestinations -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"

This command export the maximum numbers of calls for each destinations per days from the MOR database. It generates a CSV file with the specified start and end date. The CSV file contains information about day, country, calls. The generated CSV file is named with a timestamp and saved in the current working directory.`,
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
		fmt.Println("morMaxCallsNumberPerDaysByDestinations called with dateStart: " + dateStart.Format("2006-01-02 15:04:05") + " and dateEnd: " + dateEnd.Format("2006-01-02 15:04:05"))

		// Construct the SQL query with placeholders.
		request := fmt.Sprintf(`SELECT
			DATE(c.calldate) AS Day,
        	mor.destinations.name AS Destination,
        	c.prefix as Prefix,
			count(*) AS Calls 
		FROM mor.calls c inner join mor.destinations on c.prefix = mor.destinations.prefix
		WHERE 
			calldate  > '%s' AND
			calldate  < '%s' AND
			dst_device_id = 0
		GROUP BY destination,Day
		ORDER BY destination,Day;`, dateStartStr, dateEndStr)

		// Log the SQL query for debugging and tracking purposes.
		log.Print(request)

		// Send the SQL request to the MorRequest function and obtain results.
		results, err := MorRequest(request, getModelMorMaxCallsNumberPerDaysByDestinations)
		if err != nil {
			log.Fatal(err)
		}

		// Create a list to hold the results casted to the desired data model.
		var resultsCasted []ModelMorMaxCallsNumberPerDaysByDestinations
		for _, result := range results {
			resultsCasted = append(resultsCasted, result.(ModelMorMaxCallsNumberPerDaysByDestinations))
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
		fmt.Fprintln(outputFile, "Day;Country;Calls")

		// Initialize the arrey of contry and their calls
		var countryCalls []ModelMorMaxCallsNumberPerDaysByCountry

		// Process and write each result to the output file.
		for _, oneResult := range resultsCasted {
			// Process and format the prefix for phone numbers.
			oneResult.Prefix = "+" + oneResult.Prefix
			oneResult.Prefix = strings.TrimRight(oneResult.Prefix+"000000", " ")[0:6]

			// Parse and format phone number information.
			phoneNumber, err := phonenumbers.Parse(oneResult.Prefix, "")
			if err == nil {
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

				// Add the country calls to the list and/or sum the call number
				var found bool
				for i, countryCall := range countryCalls {
					if countryCall.Country == displayRegion && countryCall.Day == oneResult.Day {
						countryCalls[i].Calls += oneResult.Calls
						found = true
						break
					}
				}
				if !found {
					countryCalls = append(countryCalls, ModelMorMaxCallsNumberPerDaysByCountry{Country: displayRegion, Calls: oneResult.Calls, Day: oneResult.Day})
				}
			} else {
				// Handle the case where phone number information cannot be parsed.
				displayRegion := "UNKNOWN"

				// Add the country calls to the list and/or sum the call number
				var found bool
				for i, countryCall := range countryCalls {
					if countryCall.Country == displayRegion && countryCall.Day == oneResult.Day {
						countryCalls[i].Calls += oneResult.Calls
						found = true
						break
					}
				}
				if !found {
					countryCalls = append(countryCalls, ModelMorMaxCallsNumberPerDaysByCountry{Country: displayRegion, Calls: oneResult.Calls, Day: oneResult.Day})
				}
			}

		}

		// Write into the file
		for _, countryCall := range countryCalls {
			fmt.Fprintf(outputFile, "%s;%s;%d\n", countryCall.Day, countryCall.Country, countryCall.Calls)
		}

		// Log a message indicating the filename of the exported data.
		log.Printf("%s exported", filename)
	},
}
