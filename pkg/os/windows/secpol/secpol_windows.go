package secpol

import (
	"fmt"
	"io/ioutil"
	goos "os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

const (
	config = `[Unicode]
Unicode=yes
[Version]
signature="$CHICAGO$"
Revision=1
[Privilege Rights]
seservicelogonright = %s`
)

func getSidOfCurrentUser() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Uid, nil
}

func getSeceditPath() string {
	path, err := exec.LookPath("secedit.exe")
	if err != nil {
		logging.Debug("Cannot find 'secedit.exe' on path", err)
		return ""
	}
	return path
}

// UserAllowedToLogonAsService returns true if the user's sid is present in 'seservicelogonright' config key
func UserAllowedToLogonAsService(username string) (bool, error) {
	/* To find out the user is allowed to log on as service by security policy
	 * we export the User rights from the system security db at %windrive/defltbase.sdb using secedit.exe
	 * to a temporary file and check if the given user's SID has the user right 'seservicelogonright'
	 * see https://docs.microsoft.com/en-us/windows/security/threat-protection/security-policy-settings/security-policy-settings
	 * and https://winaero.com/blog/reset-local-security-policy-settings-all-at-once-in-windows-10/
	 */

	logonUserRight := hasLogonUserRight(username)
	if logonUserRight {
		return true, nil
	}

	return false, nil
}

func hasLogonUserRight(username string) bool {
	for _, sid := range getUsersWithLogonUserRights() {
		/* seservicelogonright could either contain the username or the SID
		 * the SID has a * at the beginning which needs to be removed to match with user's sid
		 * if its a username it doesn't have the '*' in the beginning
		 */
		userSid, err := getSidOfCurrentUser()
		if err != nil || sid == "" {
			logging.Debug("Unable to get sid of user: ", err)
		}
		if sid[1:] == userSid || sid == username {
			logging.Debug("Found user/sid in servicelogonright: ", sid)
			return true
		}
	}
	return false
}

// AllowUserToLogonAsService adds user's SID to 'seservicelogonright' config key in the security database
func AllowUserToLogonAsService(username string) error {
	userSid, err := getSidOfCurrentUser()
	if err != nil {
		return fmt.Errorf("Unable to get sid for username: %s: %v", username, err)
	}

	userLogonRightSids := getUsersWithLogonUserRights()
	if len(userLogonRightSids) == 0 {
		logging.Debugf("'seservicelogonright' config is empty or not present")
	}

	userSid = `*` + userSid

	userLogonRightSids = append(userLogonRightSids, userSid)

	seservicelogonrightConfig := fmt.Sprintf(config, strings.Join(userLogonRightSids, ","))

	return configurePolicy(seservicelogonrightConfig)

}

func getUsersWithLogonUserRights() []string {
	/* 'seservicelogon' key is a comma separated list of SIDs or usernames
	 * the users present are allowed to log on as a  service
	 * seservicelogonright = *S-1-5-21-3938078522-2608308634-3841203849-1001,*S-1-5-80-0
	 */
	seservicelogonright := getRawConfig("SeServiceLogonRight")
	logging.Debug("Got raw config from file: ", seservicelogonright)
	if len(seservicelogonright) == 0 {
		return []string{}
	}
	var sids []string
	s := strings.Split(seservicelogonright, "=")

	if len(s) > 1 {
		value := s[1]
		value = strings.TrimSpace(value)
		logging.Debug("Got users with SeServiceLogonRight: ", value)
		// the values could be more than one or just one
		// in case of just one there's no ',' and we cannot split on ','
		if strings.Contains(value, ",") {
			for _, s := range strings.Split(value, ",") {
				if len(s) != 0 {
					sids = append(sids, s)
				}
			}
			return sids
		}
		// there's only one sid/user present so return just that
		sids = append(sids, value)
	}
	return sids
}

func getRawConfig(configName string) string {
	/* security configurations are stored in a inf file
	 * each line contains a config key and a value delimited by an '=' sign
	 * the value is a list of comma separated strings
	 */

	securityInfoFilePath, err := exportConfigFromSecurityDbToFile()
	if err != nil {
		logging.Debugf("Unable to get export config: %v", err)
		return ""
	}

	defer func() {
		_ = goos.RemoveAll(filepath.Dir(securityInfoFilePath))
	}()

	// open the security config file and get the line containing the config for configName
	content, err := os.ReadFileUTF16LE(securityInfoFilePath)
	if err != nil {
		logging.Debugf("Unable to read for security info file: %v", err)
		return ""
	}

	configPattern := configName + `\s=\s.*`
	logging.Debug(configPattern)

	re := regexp.MustCompile(configPattern)
	match := re.Find(content)
	if match == nil {
		logging.Debugf("No match found, cannot find key %s in %s: %s", configName, securityInfoFilePath, string(content))
		return ""
	}

	return string(match)
}

func exportConfigFromSecurityDbToFile() (string, error) {
	// create temporary directory to export security info
	// caller should remove this directory
	path, err := ioutil.TempDir("", "crc")
	if err != nil {
		return "", err
	}

	securityInfoFilePath := filepath.Join(path, "sec.inf")
	logging.Debug("Security info file path: ", securityInfoFilePath)

	// export from security db at c:\windows\system32\defltbase.sdb
	seceditExportCmd := fmt.Sprintf("%s /export /cfg %s /areas USER_RIGHTS", getSeceditPath(), securityInfoFilePath)
	logging.Debug("Running command: ", seceditExportCmd)

	_, stdErr, err := powershell.ExecuteAsAdmin("Running secedit export command", seceditExportCmd)
	if stdErr != "" || err != nil {
		return "", fmt.Errorf("%s: %v", stdErr, err)
	}
	// wait 2 seconds for command to finish executing
	time.Sleep(2 * time.Second)
	return securityInfoFilePath, nil
}

func configurePolicy(config string) error {
	// create a config file to configure the policy using secedit
	path, err := ioutil.TempDir("", "crc")
	if err != nil {
		return err
	}
	temporarySecurityDb := filepath.Join(path, "secdef.inf")
	defer func() {
		_ = goos.RemoveAll(path)
	}()

	securityInfoFilePath := filepath.Join(path, "sec.inf")
	logging.Debug("Security info file path: ", securityInfoFilePath)
	logging.Debug("Writing config: ", config)
	err = ioutil.WriteFile(securityInfoFilePath, []byte(config), 0600)
	if err != nil {
		return err
	}
	// Write to security db at c:\windows\system32\defltbase.sdb
	seceditConfigureCmd := fmt.Sprintf(
		"%s /configure /db %s /cfg %s /areas USER_RIGHTS",
		getSeceditPath(),
		temporarySecurityDb,
		securityInfoFilePath,
	)
	logging.Debugf("Running command: %s", seceditConfigureCmd)
	_, stdErr, err := powershell.ExecuteAsAdmin("Running secedit configure command", seceditConfigureCmd)
	if stdErr != "" || err != nil {
		return fmt.Errorf("%s: %v", stdErr, err)
	}
	time.Sleep(2 * time.Second)
	return nil
}

func RemoveLogonAsServiceUserRight(username string) error {
	if !hasLogonUserRight(username) {
		logging.Debugf("User does not have log on as a service right.")
		return nil
	}
	userSid, err := getSidOfCurrentUser()
	if err != nil {
		return err
	}

	var sids []string
	for _, s := range getUsersWithLogonUserRights() {
		// skip the current user or it's SID
		// the first character in SID is '*' need to skip it
		if s[1:] == userSid || s == username {
			logging.Debugf("Found current users sid: %s, will remove", s)
			continue
		}
		sids = append(sids, s)
	}
	// write the new config
	seservicelogonrightConfig := fmt.Sprintf(config, strings.Join(sids, ","))
	return configurePolicy(seservicelogonrightConfig)
}
