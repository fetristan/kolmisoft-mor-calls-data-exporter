package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// Initialize the command.
func init() {
	rootCmd.AddCommand(morIncomingCallsDuration)
	morIncomingCallsDuration.Flags().StringP("dateStart", "s", "", "The start date of the export")
	morIncomingCallsDuration.Flags().StringP("dateEnd", "e", "", "The end date of the export")
}

// ModelMorIncomingCallsDuration represents information about incoming calls duration.
type ModelMorIncomingCallsDuration struct {
	Did         string
	Seconds     int
	Provider    string
	Username    string
	Extension   *string
	Description *string
	Status      string
	UpdateDate  string
}

// Retrieve call data from the database and return it as a slice of models.
func getModelMorIncomingCallsDuration(stmt *sql.Stmt) ([]any, error) {
	var messages []any

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var msg ModelMorIncomingCallsDuration

		if err := rows.Scan(
			&msg.Did,
			&msg.Seconds,
			&msg.Provider,
			&msg.Username,
			&msg.Extension,
			&msg.Description,
			&msg.Status,
			&msg.UpdateDate,
		); err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// Define the main Cobra command for exporting call prices.
var morIncomingCallsDuration = &cobra.Command{
	Use:   "morIncomingCallsDuration",
	Short: "Export incoming call duration data for a specified date range",
	Long: `Export incoming call duration data for a specified date range. The CSV will include the following columns: Did, Seconds, Provider, Username, Extension, Description, Status, UpdateDate, Duration (hours).

Usage:
  morIncomingCallsDuration -s [start_date] -e [end_date]

Flags:
  -s, --dateStart string   The start date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')
  -e, --dateEnd string     The end date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')

Example:
  morIncomingCallsDuration -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59"

This command export incoming call duration data for a specified date range. It generates a CSV file with the specified start and end date. The CSV will include the following columns: Did, Seconds, Provider, Username, Extension, Description, Status, UpdateDate, Duration (hours). The generated CSV file is named with a timestamp and saved in the current working directory.`,
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
		fmt.Println("morIncomingCallsDuration called with dateStart: " + dateStart.Format("2006-01-02 15:04:05") + " and dateEnd: " + dateEnd.Format("2006-01-02 15:04:05"))

		// Construct the SQL query with placeholders.
		request := fmt.Sprintf(`SELECT d.did as Did,
		IF(SUM(c.duration) IS NOT NULL, SUM(c.duration),0) as Seconds,
		p.name as Provider,
		u.username as Username,
		dv.extension as Extension,
		dv.description as Description,
		d.status as Status,
		d.closed_till as UpdateDate
		FROM mor.dids d 
		left join (SELECT dst, duration, calldate FROM mor.calls sc WHERE calldate > '%s' AND sc.calldate < '%s') c ON c.dst = d.did 
		left join mor.providers p on d.provider_id = p.id
		left join mor.users u on d.user_id = u.id 
		left join mor.devices dv on d.device_id = dv.id
		group by d.did
		order by Seconds DESC, Description;`, dateStartStr, dateEndStr)

		// Log the SQL query for debugging and tracking purposes.
		log.Print(request)

		// Send the SQL request to the MorRequest function and obtain results.
		results, err := MorRequest(request, getModelMorIncomingCallsDuration)
		if err != nil {
			log.Fatal(err)
		}

		// Create a list to hold the results casted to the desired data model.
		var resultsCasted []ModelMorIncomingCallsDuration
		for _, result := range results {
			resultsCasted = append(resultsCasted, result.(ModelMorIncomingCallsDuration))
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
		fmt.Fprintln(outputFile, "Did;Seconds;Provider;Username;Extension;Description;Status;UpdateDate;Duration (hours)")

		// Process and write each result to the output file.
		for _, oneResult := range resultsCasted {
			// Format the duration to hours, minutes, and seconds.
			durationHourMinSeconds := formatTimeSecondsToHours(oneResult.Seconds)

			// Convert the extension to a string.
			extensionStr := ""
			if oneResult.Extension != nil {
				extensionStr = *oneResult.Extension
			}

			// Convert the description to a string.
			descriptionStr := ""
			if oneResult.Description != nil {
				descriptionStr = *oneResult.Description
			}

			// Write the formatted result to the output file.
			fmt.Fprintf(outputFile, "%s;%d;%s;%s;%s;%s;%s;%s;%s\n", oneResult.Did, oneResult.Seconds, oneResult.Provider, oneResult.Username, extensionStr, descriptionStr, oneResult.Status, oneResult.UpdateDate, durationHourMinSeconds)
		}
		// Log a message indicating the filename of the exported data.
		log.Printf("%s exported", filename)
	},
}
