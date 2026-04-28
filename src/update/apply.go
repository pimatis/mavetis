package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Pimatis/mavetis/src/model"
)

type Spec struct {
	Check bool
}

var newUpdater = func() *Updater {
	return New()
}

func Run(spec Spec) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	updater := newUpdater()
	latest, err := updater.Latest(ctx)
	if err != nil {
		return err
	}
	newer, err := IsNewer(model.Version, latest)
	if err != nil {
		return err
	}
	if !newer {
	fmt.Printf("%s is already at its glorious latest version, no update required. you're running v%s; enjoy your perfectly timeless binary.\n", model.Name, model.Version)
		return nil
	}
	if spec.Check {
		fmt.Printf("update available: %s -> %s\n", model.Version, strings.TrimPrefix(latest, "v"))
		return nil
	}
	dir, err := os.MkdirTemp("", model.Name+"-update-")
	if err != nil {
		return fmt.Errorf("create update directory: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(dir)
		}
	}()
	binaryPath, err := updater.DownloadBinary(ctx, latest, dir)
	if err != nil {
		return err
	}
	executable, err := currentExecutable()
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		cleanup = false
		return scheduleWindows(binaryPath, executable, dir, latest)
	}
	if err := installUnix(binaryPath, executable); err != nil {
		return err
	}
	fmt.Printf("updated %s from %s to %s\n", model.Name, model.Version, strings.TrimPrefix(latest, "v"))
	return nil
}

func currentExecutable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	return path, nil
}

func installUnix(source string, target string) error {
	directory := filepath.Dir(target)
	if canWrite(directory) {
		if err := replaceBinary(source, target); err != nil {
			return err
		}
		return nil
	}
	if _, err := exec.LookPath("sudo"); err != nil {
		return fmt.Errorf("update requires write access to %s; rerun with sudo", directory)
	}
	command := exec.Command("sudo", "install", "-m", "0755", source, target)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("install updated binary with sudo: %w", err)
	}
	return nil
}

func replaceBinary(source string, target string) error {
	directory := filepath.Dir(target)
	temp, err := os.CreateTemp(directory, model.Name+"-swap-")
	if err != nil {
		return fmt.Errorf("create replacement file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open downloaded binary: %w", err)
	}
	defer sourceFile.Close()
	if _, err := io.Copy(temp, sourceFile); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write replacement binary: %w", err)
	}
	if err := temp.Chmod(0o755); err != nil {
		_ = temp.Close()
		return fmt.Errorf("chmod replacement binary: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close replacement binary: %w", err)
	}
	if err := os.Rename(tempPath, target); err != nil {
		return fmt.Errorf("replace executable: %w", err)
	}
	return nil
}

func canWrite(directory string) bool {
	file, err := os.CreateTemp(directory, model.Name+"-perm-")
	if err != nil {
		return false
	}
	name := file.Name()
	_ = file.Close()
	_ = os.Remove(name)
	return true
}

func scheduleWindows(source string, target string, directory string, latest string) error {
	if !canWrite(filepath.Dir(target)) {
		return fmt.Errorf("update requires an elevated PowerShell or Command Prompt for %s", target)
	}
	commandText := "ping 127.0.0.1 -n 3 > nul && copy /Y \"" + source + "\" \"" + target + "\" > nul && del /F /Q \"" + source + "\" > nul 2>&1 && rmdir /S /Q \"" + directory + "\" > nul 2>&1"
	command := exec.Command("cmd.exe", "/C", commandText)
	if err := command.Start(); err != nil {
		return fmt.Errorf("schedule windows update: %w", err)
	}
	fmt.Printf("update scheduled: %s -> %s; restart %s after the replacement finishes\n", model.Version, strings.TrimPrefix(latest, "v"), model.Name)
	return nil
}

func runtimePlatform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

func asset(platform string) (string, string, error) {
	if platform == "darwin/amd64" {
		return model.Name + "_darwin_amd64.tar.gz", model.Name, nil
	}
	if platform == "darwin/arm64" {
		return model.Name + "_darwin_arm64.tar.gz", model.Name, nil
	}
	if platform == "linux/amd64" {
		return model.Name + "_linux_amd64.tar.gz", model.Name, nil
	}
	if platform == "linux/arm64" {
		return model.Name + "_linux_arm64.tar.gz", model.Name, nil
	}
	if platform == "windows/amd64" {
		return model.Name + "_windows_amd64.zip", model.Name + ".exe", nil
	}
	if platform == "windows/arm64" {
		return model.Name + "_windows_arm64.zip", model.Name + ".exe", nil
	}
	return "", "", fmt.Errorf("unsupported platform: %s", platform)
}
