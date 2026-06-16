package network
import (
	'fmt'
	'os/exec'
)
func CreateBridge() error{
	fmt.Println('[Network ] Creating bridge md-br0')
	err :=exec.Command('ip','link','add','name','md-br0','type','bridge').Run()
	if err != nil{
		fmt.Println('[Network] may already exist')
	}
	err := exec.Command('ip','addr','add','172.20.0.1/24','dev','md-br0').Run()
	if err!= nil
	{
		fmt.Println('[Network ] IP may already exist')
	}
	err := exec.Command('ip','link','set','md-br0','up').Run()
	if err! = nil
	{
		return fmt.Errorf(F'ailed to set up Bridge: %v',err)
	}
	fmt.Println('[Network] Bridge is Ready!!')
	return nil
}
func DeleteBridge() {
	exec.Command('ip','link','delete','md-br0')
	fmt.Println('The bridge removed')
}
