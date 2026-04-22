package ctx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manager handles nuon CLI context switching via symlinks.
// Contexts are stored as individual config files under ~/.config/nuon/contexts/.
// The active context is a symlink at ~/.nuon pointing to one of these files.
type Manager struct {
	configPath   string
	contextsDir  string
	previousFile string
}

func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to determine home directory: %w", err)
	}

	contextsDir := filepath.Join(home, ".config", "nuon", "contexts")

	return &Manager{
		configPath:   filepath.Join(home, ".nuon"),
		contextsDir:  contextsDir,
		previousFile: filepath.Join(contextsDir, ".previous"),
	}, nil
}

func (m *Manager) EnsureDir() error {
	return os.MkdirAll(m.contextsDir, 0755)
}

// List returns all available context names.
func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.contextsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// Current returns the name of the currently active context by resolving the symlink.
func (m *Manager) Current() (string, error) {
	fi, err := os.Lstat(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return "", fmt.Errorf("%s is not a symlink (run 'nuon ctx -s <name>' to save it as a context first)", m.configPath)
	}

	target, err := os.Readlink(m.configPath)
	if err != nil {
		return "", err
	}

	return filepath.Base(target), nil
}

// Switch changes the active context by updating the symlink.
func (m *Manager) Switch(name string) error {
	contextFile := filepath.Join(m.contextsDir, name)
	if _, err := os.Stat(contextFile); err != nil {
		return fmt.Errorf("context %q not found", name)
	}

	// Save previous context.
	if prev, err := m.Current(); err == nil && prev != "" {
		_ = os.WriteFile(m.previousFile, []byte(prev), 0644)
	}

	// Remove existing symlink.
	if err := os.Remove(m.configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to remove %s: %w", m.configPath, err)
	}

	if err := os.Symlink(contextFile, m.configPath); err != nil {
		return fmt.Errorf("unable to create symlink: %w", err)
	}

	return nil
}

// SwitchPrevious switches to the previously active context.
func (m *Manager) SwitchPrevious() (string, error) {
	data, err := os.ReadFile(m.previousFile)
	if err != nil {
		return "", fmt.Errorf("no previous context found")
	}

	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", fmt.Errorf("no previous context found")
	}

	if err := m.Switch(name); err != nil {
		return "", err
	}

	return name, nil
}

// Unset removes the ~/.nuon symlink.
func (m *Manager) Unset() error {
	fi, err := os.Lstat(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s is not a symlink, refusing to remove", m.configPath)
	}

	return os.Remove(m.configPath)
}

// Delete removes one or more named contexts. Use "." to reference the current context.
func (m *Manager) Delete(names []string) error {
	current, _ := m.Current()

	for _, name := range names {
		if name == "." {
			if current == "" {
				return fmt.Errorf("no current context to delete")
			}
			name = current
		}

		contextFile := filepath.Join(m.contextsDir, name)
		if _, err := os.Stat(contextFile); err != nil {
			return fmt.Errorf("context %q not found", name)
		}

		if err := os.Remove(contextFile); err != nil {
			return fmt.Errorf("unable to delete context %q: %w", name, err)
		}

		// If we deleted the active context, remove the symlink.
		if name == current {
			_ = m.Unset()
		}

		fmt.Fprintf(os.Stderr, "Deleted context %q\n", name)
	}

	return nil
}

// Rename renames a context. Use "." as oldName to reference the current context.
func (m *Manager) Rename(oldName, newName string) error {
	current, _ := m.Current()

	if oldName == "." {
		if current == "" {
			return fmt.Errorf("no current context to rename")
		}
		oldName = current
	}

	oldPath := filepath.Join(m.contextsDir, oldName)
	newPath := filepath.Join(m.contextsDir, newName)

	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("context %q not found", oldName)
	}

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("context %q already exists", newName)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("unable to rename: %w", err)
	}

	// If the renamed context was active, update the symlink.
	if oldName == current {
		_ = os.Remove(m.configPath)
		_ = os.Symlink(newPath, m.configPath)
	}

	return nil
}

// Save stores a config file as a named context.
// If srcFile is empty, it saves the current ~/.nuon and replaces it with a symlink.
// If srcFile is provided, it copies that file into contexts/<name> without modifying ~/.nuon.
func (m *Manager) Save(name, srcFile string) error {
	if err := m.EnsureDir(); err != nil {
		return err
	}

	contextFile := filepath.Join(m.contextsDir, name)
	if _, err := os.Stat(contextFile); err == nil {
		return fmt.Errorf("context %q already exists", name)
	}

	if srcFile != "" {
		// Save an external file as a named context.
		data, err := os.ReadFile(srcFile)
		if err != nil {
			return fmt.Errorf("unable to read %s: %w", srcFile, err)
		}
		return os.WriteFile(contextFile, data, 0644)
	}

	// Save the current ~/.nuon as a named context.
	fi, err := os.Lstat(m.configPath)
	if err != nil {
		return fmt.Errorf("%s does not exist", m.configPath)
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("unable to read config: %w", err)
	}
	if err := os.WriteFile(contextFile, data, 0644); err != nil {
		return err
	}

	// If it was a regular file, remove it so we can replace with a symlink.
	if fi.Mode()&os.ModeSymlink == 0 {
		if err := os.Remove(m.configPath); err != nil {
			return fmt.Errorf("unable to remove original config: %w", err)
		}
	}

	// Point ~/.nuon at the new context.
	_ = os.Remove(m.configPath)
	if err := os.Symlink(contextFile, m.configPath); err != nil {
		return fmt.Errorf("unable to create symlink: %w", err)
	}

	return nil
}
