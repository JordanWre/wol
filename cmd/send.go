package cmd

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.Flags().StringP("mac", "m", "", "MAC address of the device to wake up")
	sendCmd.Flags().StringP("name", "n", "", "Name of the device to wake up")
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a magic packet to specified mac address",
	Long:  "Send a magic packet to wake up a device on the network using the specified mac address",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Only one of the flags should be specified
		if cmd.Flags().Changed("mac") == cmd.Flags().Changed("name") {
			return fmt.Errorf("either --mac or --name must be specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var mac net.HardwareAddr
		var err error

		// Retrieve mac address using one of the flags
		switch true {
		case cmd.Flags().Changed("mac"):
			value, err := cmd.Flags().GetString("mac")
			if err != nil {
				cobra.CheckErr(err)
			}

			mac, err = net.ParseMAC(value)
			if err != nil {
				cobra.CheckErr(err)
			}
		case cmd.Flags().Changed("name"):
			// Get the name of the machine
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				cobra.CheckErr(err)
			}

			// Find machine with the specified name
			for _, machine := range cfg.Machines {
				if strings.EqualFold(machine.Name, name) {
					mac, err = net.ParseMAC(machine.Mac)
					if err != nil {
						cobra.CheckErr(err)
					}
				}
			}

			if mac == nil {
				cobra.CheckErr(fmt.Errorf("machine with name %q not found", name))
			}
		default:
			log.Fatalf("mac address should come from either --mac or --name")
		}

		log.Printf("Sending magic packet to %s", mac)
		err = sendMagicPacket(mac)
		if err != nil {
			cobra.CheckErr(err)
		}

		log.Printf("Magic packet sent")
	},
}

func sendMagicPacket(mac net.HardwareAddr) error {
	// Build magic packet
	// Create a buffer for the magic packet
	packet := make([]byte, 102)
	// Set the synchronization stream (first 6 bytes are 0xFF)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	// Copy the MAC address 16 times into the packet
	for i := 1; i <= 16; i++ {
		copy(packet[i*6:], mac)
	}

	// Broadcast magic packet
	addr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 9,
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}
