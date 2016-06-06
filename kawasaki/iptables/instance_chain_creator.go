package iptables

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/pivotal-golang/lager"
)

type InstanceChainCreator struct {
	iptables *IPTablesController
}

func NewInstanceChainCreator(iptables *IPTablesController) *InstanceChainCreator {
	return &InstanceChainCreator{
		iptables: iptables,
	}
}

func (cc *InstanceChainCreator) Create(logger lager.Logger, instanceId, bridgeName string, ip net.IP, network *net.IPNet) error {
	instanceChain := cc.iptables.InstanceChain(instanceId)

	if err := cc.iptables.CreateChain("nat", instanceChain); err != nil {
		return err
	}

	// Bind nat instance chain to nat prerouting chain
	cmd := exec.Command("iptables", "--wait", "--table", "nat", "-A", cc.iptables.preroutingChain, "--jump", instanceChain)
	if err := cc.iptables.run("create-instance-chains", cmd); err != nil {
		return err
	}

	// Enable NAT for traffic coming from containers
	cmd = exec.Command("sh", "-c", fmt.Sprintf(
		`(iptables --wait --table nat -S %s | grep "\-j MASQUERADE\b" | grep -q -F -- "-s %s") || iptables --wait --table nat -A %s --source %s ! --destination %s --jump MASQUERADE`,
		cc.iptables.postroutingChain, network.String(), cc.iptables.postroutingChain,
		network.String(), network.String(),
	))
	if err := cc.iptables.run("create-instance-chains", cmd); err != nil {
		return err
	}

	// Create filter instance chain
	if err := cc.iptables.CreateChain("filter", instanceChain); err != nil {
		return err
	}

	// Allow intra-subnet traffic (Linux ethernet bridging goes through ip stack)
	cmd = exec.Command("iptables", "--wait", "-A", instanceChain, "-s", network.String(), "-d", network.String(), "-j", "ACCEPT")
	if err := cc.iptables.run("create-instance-chains", cmd); err != nil {
		return err
	}

	// Otherwise, use the default filter chain
	cmd = exec.Command("iptables", "--wait", "-A", instanceChain, "--goto", cc.iptables.defaultChain)
	if err := cc.iptables.run("create-instance-chains", cmd); err != nil {
		return err
	}

	// Bind filter instance chain to filter forward chain
	cmd = exec.Command("iptables", "--wait", "-I", cc.iptables.forwardChain, "2", "--in-interface", bridgeName, "--source", ip.String(), "--goto", instanceChain)
	if err := cc.iptables.run("create-instance-chains", cmd); err != nil {
		return err
	}

	// Create Logging Chain
	return cc.createLoggingChain(logger, instanceId)
}

func (cc *InstanceChainCreator) createLoggingChain(logger lager.Logger, instanceId string) error {
	instanceChain := cc.iptables.InstanceChain(instanceId)
	loggingChain := fmt.Sprintf("%s-log", instanceChain)

	if err := cc.iptables.CreateChain("filter", loggingChain); err != nil {
		return err
	}

	cmd := exec.Command("iptables", "--wait", "-A", loggingChain, "-m", "conntrack", "--ctstate", "NEW,UNTRACKED,INVALID", "--protocol", "tcp", "--jump", "LOG", "--log-prefix", instanceId)
	if err := cc.iptables.run("create-instance-chains", cmd); err != nil {
		return err
	}

	cmd = exec.Command("iptables", "--wait", "-A", loggingChain, "--jump", "RETURN")
	if err := cc.iptables.run("create-instance-chains", cmd); err != nil {
		return err
	}

	return nil
}

func (cc *InstanceChainCreator) Destroy(logger lager.Logger, instanceId string) error {
	instanceChain := cc.iptables.InstanceChain(instanceId)

	// Prune nat prerouting chain
	cmd := exec.Command("sh", "-c", fmt.Sprintf(
		`iptables --wait --table nat -S %s 2> /dev/null | grep "\-j %s\b" | sed -e "s/-A/-D/" | xargs --no-run-if-empty --max-lines=1 iptables --wait --table nat`,
		cc.iptables.preroutingChain, instanceChain,
	))
	if err := cc.iptables.run("prune-prerouting-chain", cmd); err != nil {
		return err
	}

	// Flush instance chain
	if err := cc.iptables.FlushChain("nat", instanceChain); err != nil {
		return err
	}

	// Delete nat instance chain
	if err := cc.iptables.DeleteChain("nat", instanceChain); err != nil {
		return err
	}

	// Prune forward chain
	cmd = exec.Command("sh", "-c", fmt.Sprintf(
		`iptables --wait -S %s 2> /dev/null | grep "\-g %s\b" | sed -e "s/-A/-D/" | xargs --no-run-if-empty --max-lines=1 iptables --wait`,
		cc.iptables.forwardChain, instanceChain,
	))
	if err := cc.iptables.run("prune-forward-chain", cmd); err != nil {
		return err
	}

	// Flush instance chain
	cc.iptables.FlushChain("filter", instanceChain)

	// delete instance chain
	cc.iptables.DeleteChain("filter", instanceChain)

	// delete the logging chain
	instanceLoggingChain := fmt.Sprintf("%s-log", instanceChain)
	cc.iptables.FlushChain("filter", instanceLoggingChain)
	cc.iptables.DeleteChain("filter", instanceLoggingChain)

	return nil
}
