package network
import (
	'fmt'
	'os/exec'
)
func SetupVeth(containerPID int) error{
	fmt.Printf('[Network] setting up virtual cable for container PID %d \n',containerPID)
	err := exec.Command('ip','link','add','veth_host','type','veth','peer','name','veth_child').Run()
	if err != nil{
		return fmt.Errorf('failed to create  veth pair: %v',err)
	}
	err = exec.Command("ip", "link", "set", "veth_host", "master", "md-br0").Run()
    if err != nil {
        return fmt.Errorf("failed to attach veth_host to bridge: %v", err)
    }
	err = exec.Command("ip", "link", "set", "veth_host", "up").Run()
    if err != nil {
        return fmt.Errorf("failed to bring veth_host up: %v", err)
    }
	pidStr := fmt.Sprintf("%d", containerPID)
    err = exec.Command("ip", "link", "set", "veth_child", "netns", pidStr).Run()
    if err != nil {
        return fmt.Errorf("failed to move veth_child into container: %v", err)
    }
	err = exec.Command(
        "nsenter", "-t", pidStr, "-n",
        "ip", "addr", "add", "172.20.0.2/24", "dev", "veth_child",
    ).Run()
    if err != nil {
        return fmt.Errorf("failed to assign IP inside container: %v", err)
    }
	err = exec.Command(
        "nsenter", "-t", pidStr, "-n",
        "ip", "link", "set", "veth_child", "up",
    ).Run()
    if err != nil {
        return fmt.Errorf("failed to bring veth_child up: %v", err)
    }
	err = exec.Command(
        "nsenter", "-t", pidStr, "-n",
        "ip", "route", "add", "default", "via", "172.20.0.1",
    ).Run()
    if err != nil {
        return fmt.Errorf("failed to add default route in container: %v", err)
    }

    fmt.Println("[Network] Virtual cable connected successfully!")
    return nil
	func CleanupVeth() {
    exec.Command("ip", "link", "del", "veth_host").Run()
    fmt.Println("[Network] Virtual cable removed.")
}