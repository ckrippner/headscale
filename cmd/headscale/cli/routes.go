package cli

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	v1 "github.com/juanfont/headscale/gen/go/headscale/v1"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(routesCmd)

	listRoutesCmd.Flags().Uint64P("identifier", "i", 0, "Node identifier (ID)")
	err := listRoutesCmd.MarkFlagRequired("identifier")
	if err != nil {
		log.Fatalf(err.Error())
	}
	routesCmd.AddCommand(listRoutesCmd)

	enableRouteCmd.Flags().StringSliceP("route", "r", []string{}, "List (or repeated flags) of routes to enable")
	enableRouteCmd.Flags().Uint64P("identifier", "i", 0, "Node identifier (ID)")
	err = enableRouteCmd.MarkFlagRequired("identifier")
	if err != nil {
		log.Fatalf(err.Error())
	}

	routesCmd.AddCommand(enableRouteCmd)

	nodeCmd.AddCommand(routesCmd)
}

var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Manage the routes of Headscale",
}

var listRoutesCmd = &cobra.Command{
	Use:   "list",
	Short: "List routes advertised and enabled by a given node",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")

		machineId, err := cmd.Flags().GetUint64("identifier")
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Error getting machine id from flag: %s", err), output)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, conn := getHeadscaleGRPCClient(ctx)
		defer conn.Close()

		request := &v1.GetMachineRouteRequest{
			MachineId: machineId,
		}

		response, err := client.GetMachineRoute(ctx, request)
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Cannot get nodes: %s", err), output)
			return
		}

		if output != "" {
			SuccessOutput(response.Routes, "", output)
			return
		}

		d := routesToPtables(response.Routes)
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Error converting to table: %s", err), output)
			return
		}

		err = pterm.DefaultTable.WithHasHeader().WithData(d).Render()
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Failed to render pterm table: %s", err), output)
			return
		}
	},
}

var enableRouteCmd = &cobra.Command{
	Use:   "enable",
	Short: "Set the enabled routes for a given node",
	Long: `This command will take a list of routes that will _replace_ 
the current set of routes on a given node.
If you would like to disable a route, simply run the command again, but 
omit the route you do not want to enable.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		machineId, err := cmd.Flags().GetUint64("identifier")
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Error getting machine id from flag: %s", err), output)
			return
		}

		routes, err := cmd.Flags().GetStringSlice("route")
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Error getting routes from flag: %s", err), output)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, conn := getHeadscaleGRPCClient(ctx)
		defer conn.Close()

		request := &v1.EnableMachineRoutesRequest{
			MachineId: machineId,
			Routes:    routes,
		}

		response, err := client.EnableMachineRoutes(ctx, request)
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Cannot register machine: %s\n", err), output)
			return
		}

		if output != "" {
			SuccessOutput(response.Routes, "", output)
			return
		}

		d := routesToPtables(response.Routes)
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Error converting to table: %s", err), output)
			return
		}

		err = pterm.DefaultTable.WithHasHeader().WithData(d).Render()
		if err != nil {
			ErrorOutput(err, fmt.Sprintf("Failed to render pterm table: %s", err), output)
			return
		}
	},
}

// routesToPtables converts the list of routes to a nice table
func routesToPtables(routes *v1.Routes) pterm.TableData {
	d := pterm.TableData{{"Route", "Enabled"}}

	for _, route := range routes.GetAdvertisedRoutes() {
		enabled := isStringInSlice(routes.EnabledRoutes, route)

		d = append(d, []string{route, strconv.FormatBool(enabled)})
	}
	return d
}

func isStringInSlice(strs []string, s string) bool {
	for _, s2 := range strs {
		if s == s2 {
			return true
		}
	}

	return false
}
