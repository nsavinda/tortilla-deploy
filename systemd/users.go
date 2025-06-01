package systemd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"AutoPuller/config"
)

type RunAs struct {
	User  string
	Group string
}

// Removed duplicate ServiceConfig struct; use config.ServiceConfig from imported package

func ensureGroupExists(groupname string) error {
	if groupname == "" {
		return nil
	}
	_, err := user.LookupGroup(groupname)
	if err == nil {
		return nil // group exists
	}

	// cmd := exec.Command("groupadd", groupname) with directory creation
	cmd := exec.Command("groupadd", "-r", groupname)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("groupadd %s failed: %v - %s", groupname, err, output)
	}
	log.Printf("Created group: %s", groupname)
	return nil
}

func ensureUserExists(username, groupname string) error {
	if username == "" {
		return nil
	}

	_, err := user.Lookup(username)
	if err == nil {
		return nil // user exists
	}

	args := []string{"-m", "-s", "/bin/bash"}
	if groupname != "" {
		args = append(args, "-g", groupname)
	}
	args = append(args, username)

	cmd := exec.Command("useradd", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("useradd %s failed: %v - %s", username, err, output)
	}
	log.Printf("Created user: %s", username)
	return nil
}

func CreateAllUsers() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	for serviceName := range cfg.Services {
		svc, err := cfg.GetService(serviceName)
		if err != nil {
			return fmt.Errorf("failed to get service %q: %w", serviceName, err)
		}
		if svc.RunAs.User == "" && svc.RunAs.Group == "" {
			continue // nothing to do for this service
		}

		log.Printf("Setting up user/group for service: %s", svc.Name)

		if svc.RunAs.Group != "" {
			if err := ensureGroupExists(svc.RunAs.Group); err != nil {
				return fmt.Errorf("failed to ensure group %q for service %q: %w", svc.RunAs.Group, svc.Name, err)
			}
		}

		if svc.RunAs.User != "" {
			if err := ensureUserExists(svc.RunAs.User, svc.RunAs.Group); err != nil {
				return fmt.Errorf("failed to ensure user %q for service %q: %w", svc.RunAs.User, svc.Name, err)
			}
		}
	}

	return nil
}

func RunAsUserGroup(runAs RunAs, cmd *exec.Cmd) (*exec.Cmd, error) {
	if runAs.User == "" && runAs.Group == "" {
		return cmd, nil // no user/group specified
	}

	fmt.Printf("Running command as user: %s, group: %s\n", runAs.User, runAs.Group)

	cred := &syscall.Credential{}

	if runAs.User != "" {
		userInfo, err := user.Lookup(runAs.User)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup user %q: %w", runAs.User, err)
		}
		uid, _ := strconv.Atoi(userInfo.Uid)
		cred.Uid = uint32(uid)
	}

	if runAs.Group != "" {
		groupInfo, err := user.LookupGroup(runAs.Group)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup group %q: %w", runAs.Group, err)
		}
		gid, _ := strconv.Atoi(groupInfo.Gid)
		cred.Gid = uint32(gid)
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf("USER=%s", runAs.User))
	cmd.Env = append(cmd.Env, fmt.Sprintf("GROUP=%s", runAs.Group))
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=/home/%s", runAs.User))

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: cred,
	}

	return cmd, nil
}

func SetDirOwnership(dir string, runAs RunAs) error {
	if runAs.User == "" && runAs.Group == "" {
		return nil // no ownership change needed
	}

	fmt.Printf("Setting ownership of directory %s to user: %s, group: %s\n", dir, runAs.User, runAs.Group)

	// if dir does not exist, create it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// permissions
	if err := os.Chmod(dir, 0777); err != nil {
		return fmt.Errorf("failed to set permissions for %s: %w", dir, err)
	}

	cmd := exec.Command("chown", "-R", fmt.Sprintf("%s:%s", runAs.User, runAs.Group), dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set ownership for %s: %w", dir, err)
	}

	return nil
}

func ChownR(path string, runAs RunAs) error {
	u, err := user.Lookup(runAs.User)
	if err != nil {
		return err
	}
	g, err := user.LookupGroup(runAs.Group)
	if err != nil {
		return err
	}

	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(g.Gid)

	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(p, uid, gid)
	})
}
