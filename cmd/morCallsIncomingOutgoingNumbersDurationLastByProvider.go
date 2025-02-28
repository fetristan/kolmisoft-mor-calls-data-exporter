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
	rootCmd.AddCommand(morCallsIncomingOutgoingNumbersDurationLastByProvider)
	morCallsIncomingOutgoingNumbersDurationLastByProvider.Flags().StringP("provider", "p", "", "A part of the provider name of the export")
	morCallsIncomingOutgoingNumbersDurationLastByProvider.Flags().StringP("dateStart", "s", "", "The start date of the export")
	morCallsIncomingOutgoingNumbersDurationLastByProvider.Flags().StringP("dateEnd", "e", "", "The end date of the export")
}

// ModelMorCallsIncomingOutgoingNumbersDurationLastByProvider represents information about incoming calls duration.
type ModelMorCallsIncomingOutgoingNumbersDurationLastByProvider struct {
	Did              string
	IncomingCalls    int
	IncomingDuration int
	LastIncoming     *string
	OutgoingCalls    int
	OutgoingDuration int
	LastOutgoing     *string
	Provider         string
}

// Retrieve call data from the database and return it as a slice of models.
func getModelMorCallsIncomingOutgoingNumbersDurationLastByProvider(stmt *sql.Stmt) ([]any, error) {
	var messages []any

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var msg ModelMorCallsIncomingOutgoingNumbersDurationLastByProvider

		if err := rows.Scan(
			&msg.Did,
			&msg.IncomingCalls,
			&msg.IncomingDuration,
			&msg.LastIncoming,
			&msg.OutgoingCalls,
			&msg.OutgoingDuration,
			&msg.LastOutgoing,
			&msg.Provider,
		); err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// Define the main Cobra command for exporting call prices.
var morCallsIncomingOutgoingNumbersDurationLastByProvider = &cobra.Command{
	Use:   "morCallsIncomingOutgoingNumbersDurationLastByProvider",
	Short: "Export incoming and outgoing calls data for actives numbers (calls numbers, duration,last call date) for a specified date range and provider",
	Long: `Export incoming and outgoing calls data for actives numbers  (calls numbers, duration,last call date) for a specified date range and provider. The CSV will include the following columns: DID, Incoming Calls, Incoming Duration (seconds), Last Incoming, Outgoing Calls, Outgoing Duration (seconds), Last Outgoing, Provider.

Usage:
  morCallsIncomingOutgoingNumbersDurationLastByProvider -s [start_date] -e [end_date] -p [provider]

Flags:
  -p, --provider  string   A part of the provider name of the export (e.g., 'sfr').
  -s, --dateStart string   The start date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')
  -e, --dateEnd   string   The end date of the export (e.g., 'YYYY-MM-DD HH:mm:SS')

Example:
  morCallsIncomingOutgoingNumbersDurationLastByProvider -s "2023-01-01 00:00:00" -e "2023-01-31 23:59:59" -p "sfr"

Export incoming and outgoing calls data for actives numbers (calls numbers, duration,last call date) for a specified date range and provider. The CSV will include the following columns: DID, Incoming Calls, Incoming Duration (seconds), Last Incoming, Outgoing Calls, Outgoing Duration (seconds), Last Outgoing, Provider. The generated CSV file is named with a timestamp and saved in the current working directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Obtain the start and end date strings from the command-line flags.
		dateStartStr, _ := cmd.Flags().GetString("dateStart")
		dateEndStr, _ := cmd.Flags().GetString("dateEnd")
		provider := cmd.Flag("provider").Value.String()

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
		fmt.Println("morCallsIncomingOutgoingNumbersDurationLastByProvider called with dateStart: " + dateStart.Format("2006-01-02 15:04:05") + " and dateEnd: " + dateEnd.Format("2006-01-02 15:04:05") + " and provider: " + provider)

		// Construct the SQL query with placeholders.
		request := fmt.Sprintf(`SELECT
	d.did AS DID,
	COUNT(DISTINCT CASE WHEN c.provider_id = 0 THEN c.id END) AS IncomingCalls,
	SUM(CASE WHEN c.provider_id = 0 THEN c.billsec ELSE 0 END) AS IncomingDuration,
	MAX(CASE WHEN c.provider_id = 0 THEN c.calldate END) AS LastIncoming,
	COUNT(DISTINCT CASE WHEN c.provider_id != 0 THEN c.id END) AS OutgoingCalls,
	SUM(CASE WHEN c.provider_id != 0 THEN c.billsec ELSE 0 END) AS OutgoingDuration,
	MAX(CASE WHEN c.provider_id != 0 THEN c.calldate END) AS LastOutgoing,
	p.name AS Provider
	FROM
		dids d
	LEFT JOIN
		calls c ON (c.dst = d.did OR c.src = d.did)
		AND c.calldate > '%s'
		AND c.calldate < '%s'
	LEFT JOIN
		providers p ON (d.provider_id = p.id)
	WHERE
		d.status = 'active'
		AND p.name LIKE '%%%s%%'
	GROUP BY
		d.did, d.provider_id
	ORDER BY
		d.did;`, dateStartStr, dateEndStr, provider)

		// Log the SQL query for debugging and tracking purposes.
		log.Print(request)

		// Send the SQL request to the MorRequest function and obtain results.
		results, err := MorRequest(request, getModelMorCallsIncomingOutgoingNumbersDurationLastByProvider)
		if err != nil {
			log.Fatal(err)
		}

		// Create a list to hold the results casted to the desired data model.
		var resultsCasted []ModelMorCallsIncomingOutgoingNumbersDurationLastByProvider
		for _, result := range results {
			resultsCasted = append(resultsCasted, result.(ModelMorCallsIncomingOutgoingNumbersDurationLastByProvider))
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
		fmt.Fprintln(outputFile, "DID;Incoming Calls;Incoming Duration (seconds);Last Incoming;Outgoing Calls;Outgoing Duration (seconds);Last Outgoing;Provider")

		// Process and write each result to the output file.
		for _, oneResult := range resultsCasted {
			LastIncomingStr := ""
			if oneResult.LastIncoming != nil {
				LastIncomingStr = *oneResult.LastIncoming
			}
			LastOutgoingStr := ""
			if oneResult.LastOutgoing != nil {
				LastOutgoingStr = *oneResult.LastOutgoing
			}
			// Write the formatted result to the output file.
			fmt.Fprintf(outputFile, "%s;%d;%d;%s;%d;%d;%s;%s\n", oneResult.Did, oneResult.IncomingCalls, oneResult.IncomingDuration, LastIncomingStr, oneResult.OutgoingCalls, oneResult.OutgoingDuration, LastOutgoingStr, oneResult.Provider)
		}
		// Log a message indicating the filename of the exported data.
		log.Printf("%s exported", filename)
	},
}
