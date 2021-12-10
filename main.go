package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"util"
)

const userCredsFile = "dhc_collibra.json"

func main() {
	vars, err := loadUserCreds()
	if err != nil {
		log.Fatal(err)
		return
	}

	backups, err := loadAvailableBackups(vars)
	if err != nil {
		log.Fatal(err)
		return
	}

	lastBackup := getYeserdayBackup(backups, vars)

	err = downloadBackup(lastBackup, vars)
	if err != nil {
		log.Fatal(err)
		return
	}
}

type collibraVars struct {
	DGC           string `json:"dgc"`
	UserName      string `json:"username"`
	Password      string `json:"password""`
	EncryptionKey string `json:"encryption-key"`
	BackupDir     string `json:"backup-dir"`
	BackupFormat  string `json:"backup-format"`
}

type backup struct {
	CreatedDate       float64 `json:"createdDate,omitempty"`
	ModifiedDate      float64 `json:"modifiedDate,omitempty"`
	Id                string  `json:"id"`
	BackupInformation struct {
		Name                  string   `json:"name"`
		Description           string   `json:"description,omitempty"`
		AppVersion            string   `json:"appVersion,omitempty"`
		CreatedByEmail        string   `json:"createdByEmail,omitempty"`
		Date                  float64  `json:"date,omitempty"`
		BackupSpecificationId string   `json:"backupSpecificationId,omitempty"`
		EnvironmentId         string   `json:"environmentId,omitempty"`
		DGCBackupOptions      []string `json:"dgcBackupOptions,omitempty"`
		RepoBackupOptions     []string `json:"repoBackupOptions,omitempty"`
	}
	StepStateMap map[string]map[string]string `json:"stepStateMap,omitempty"`
	Size         int64                        `json:"size,omitempty"`
}

func loadUserCreds() (collibraVars, error) {
	user, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	varsFile := fmt.Sprintf("%s/.creds/%s", user, userCredsFile)
	util.PrintHeader(fmt.Sprintf("Loading Creds from %v", varsFile))

	vars, err := ioutil.ReadFile(varsFile)
	var collibra collibraVars
	if err != nil {
		return collibra, err
	}

	err = json.Unmarshal(vars, &collibra)
	if err != nil {
		return collibra, err
	}

	return collibra, error(nil)
}

func basicAuth(vars collibraVars) string {
	auth := vars.UserName + ":" + vars.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func loadAvailableBackups(vars collibraVars) ([]backup, error) {
	util.PrintHeader("Loading available backups")
	var availableBackups []backup

	// Calling /rest/backup
	dgcBackupURI := fmt.Sprintf("%s/rest/backup", vars.DGC)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", dgcBackupURI, nil)
	req.Header.Add("Authorization", "Basic "+basicAuth(vars))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return availableBackups, err
	}

	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &availableBackups)
	fmt.Printf("==> Number of backups available: %d", len(body))
	return availableBackups, err
}

func getYeserdayBackup(backups []backup, vars collibraVars) backup {
	yDate := time.Now()
	yString := fmt.Sprintf("%d-%02d-%02d", yDate.Year(), yDate.Month(), yDate.Day())
	dateToSearch := strings.Replace(vars.BackupFormat, "<DATE>", yString, 1)

	util.PrintHeader(fmt.Sprintf("Finding yeserday's backup, based off format: %s", dateToSearch))

	for _, backup := range backups {
		if backup.BackupInformation.Name == dateToSearch {
			return backup
		}
	}
	return backup{}
}

func downloadBackup(backup backup, vars collibraVars) error {
	util.PrintHeader(fmt.Sprintf("Downloading Backup: %v", backup.BackupInformation.Name))
	dgcBackupURI := fmt.Sprintf("%s/rest/backup/file/%s", vars.DGC, backup.Id)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", dgcBackupURI, strings.NewReader(fmt.Sprintf("key=%v", vars.EncryptionKey)))
	req.Header.Add("Authorization", "Basic "+basicAuth(vars))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	backupFile := fmt.Sprintf("%s/%s.zip", vars.BackupDir, backup.BackupInformation.Name)
	file, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("==> File Downloaded, loc: %s\n", backupFile)
	fmt.Printf("==> File Downloaded, size %d\n", size)
	defer file.Close()
	return error(nil)
}
