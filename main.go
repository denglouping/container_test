package main

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}
		args := context.Args()
		tty := context.Bool("ti")
		Run(tty, args)
		return nil
	},
}

func NewParentProcess(tty bool, args []string) *exec.Cmd {
	execArgs := []string{"init"}
	execArgs = append(execArgs, args...)
	cmd := exec.Command("/proc/self/exe", execArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd
}

func Run(tty bool, args []string) {
	parent := NewParentProcess(tty, args)
	if err := parent.Start(); err != nil {
		klog.Fatalf(err.Error())
	}
	//err := parent.Wait()
	//if err != nil {
	//	klog.Fatalf(err.Error())
	//}
	//os.Exit(-1)
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}
		err := RunContainerInitProcess(context.Args())
		return err
	},
}

func RunContainerInitProcess(args []string) error {
	ioutil.WriteFile("log1", []byte(fmt.Sprintf("start cmd: %s", args)), 0644)

	err := setUpMount()
	if err != nil {
		ioutil.WriteFile("log2", []byte("setUpMount err : "+err.Error()), 0644)
	}

	if err := syscall.Exec("/bin/"+args[0], args, os.Environ()); err != nil {
		ioutil.WriteFile("log4", []byte("exec err : "+err.Error()), 0644)
		klog.Fatalf(err.Error())
	}
	ioutil.WriteFile("/data/go-code/container_test/log5", []byte("done"), 0644)
	return nil
}

func pivotRoot(root string) error {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount bind failed: %s", err.Error())
	}

	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("Mkdir failed: %s", err.Error())
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("PivotRoot failed: %s", err.Error())
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("Chdir failed: %s", err.Error())
	}

	pivotDir = filepath.Join("/", ".pivot_root")

	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("Unmount failed: %s", err.Error())
	}

	return os.Remove(pivotDir)

}

func setUpMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		klog.Error(err)
		return err
	}

	err = pivotRoot(filepath.Join(pwd, "busybox"))
	if err != nil {
		klog.Error(err)
		return err
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		err := ioutil.WriteFile("log2", []byte("mount err : "+err.Error()), 0644)
		if err != nil {
			fmt.Println("写入文件失败:", err)
		}
		klog.Fatalf(err.Error())
	}

	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		err := ioutil.WriteFile("log2", []byte("mount err : "+err.Error()), 0644)
		if err != nil {
			fmt.Println("写入文件失败:", err)
		}
		klog.Fatalf(err.Error())
	}
	return err

}

// 还需要修改环境变量

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = ""
	app.Commands = []cli.Command{
		initCommand,
		runCommand,
	}

	app.Before = func(context *cli.Context) error {

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		klog.Error(err.Error())
	}
}
