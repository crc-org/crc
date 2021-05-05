package goautoit

import "log"

//IsAdmin -- Checks if the current user has full administrator privileges.
func IsAdmin() int {
	ret, _, lastErr := isAdmin.Call()
	if ret == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ProcessClose -- Terminates a named process.
func ProcessClose(process string) int {
	ret, _, lastErr := processClose.Call(strPtr(process))
	if ret == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ProcessExists -- Checks to see if a specified process exists.
func ProcessExists(process string) int {
	ret, _, lastErr := processExists.Call(strPtr(process))
	if ret == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ProcessSetPriority -- Changes the priority of a process.
func ProcessSetPriority(process string, priority int) int {
	ret, _, lastErr := processSetPriority.Call(strPtr(process), intPtr(priority))
	if ret == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ProcessWait -- Pauses script execution until a given process exists.
func ProcessWait(process string, args ...interface{}) int {
	var timeout int
	var ok bool

	if len(args) == 0 {
		timeout = 0
	} else if len(args) == 1 {
		if timeout, ok = args[0].(int); !ok {
			panic("timeout must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := processWait.Call(strPtr(process), intPtr(timeout))
	if ret == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

//ProcessWaitClose -- Pauses script execution until a given process does not exist.
func ProcessWaitClose(process string, args ...interface{}) int {
	var timeout int
	var ok bool

	if len(args) == 0 {
		timeout = 0
	} else if len(args) == 1 {
		if timeout, ok = args[0].(int); !ok {
			panic("timeout must be a int")
		}
	} else {
		panic("Error parameters")
	}
	ret, _, lastErr := processWaitClose.Call(strPtr(process), intPtr(timeout))
	if ret == 0 {
		log.Println(lastErr)
	}
	return int(ret)
}

// RunWait -- Runs an external program and pauses script execution until the program finishes.
// flag 3(max) 6(min) 9(normal) 0(hide)
func RunWait(szProgram string, args ...interface{}) int {
	var szDir string
	var flag int
	var ok bool
	if len(args) == 0 {
		szDir = ""
		flag = SWShowNormal
	} else if len(args) == 1 {
		if szDir, ok = args[0].(string); !ok {
			panic("szDir must be a string")
		}
		flag = SWShowNormal
	} else if len(args) == 2 {
		if szDir, ok = args[0].(string); !ok {
			panic("szDir must be a string")
		}
		if flag, ok = args[1].(int); !ok {
			panic("flag must be a int")
		}
	} else {
		panic("Too more parameter")
	}
	pid, _, lastErr := runWait.Call(strPtr(szProgram), strPtr(szDir), intPtr(flag))
	// log.Println(pid)
	if int(pid) == 0 {
		log.Println(lastErr)
	}
	return int(pid)
}

//RunAs -- Runs an external program under the context of a different user.
func RunAs(user, domain, password string, loginFlag int, szProgram string, args ...interface{}) int {
	var szDir string
	var flag int
	var ok bool
	if len(args) == 0 {
		szDir = ""
		flag = SWShowNormal
	} else if len(args) == 1 {
		if szDir, ok = args[0].(string); !ok {
			panic("szDir must be a string")
		}
		flag = SWShowNormal
	} else if len(args) == 2 {
		if szDir, ok = args[0].(string); !ok {
			panic("szDir must be a string")
		}
		if flag, ok = args[1].(int); !ok {
			panic("flag must be a int")
		}
	} else {
		panic("Too more parameter")
	}
	pid, _, lastErr := runAs.Call(strPtr(user), strPtr(domain), strPtr(password), intPtr(loginFlag), strPtr(szProgram), strPtr(szDir), intPtr(flag))
	// log.Println(pid)
	if int(pid) == 0 {
		log.Println(lastErr)
	}
	return int(pid)
}

//RunAsWait -- Runs an external program under the context of a different user and pauses script execution until the program finishes.
func RunAsWait(user, domain, password string, loginFlag int, szProgram string, args ...interface{}) int {
	var szDir string
	var flag int
	var ok bool
	if len(args) == 0 {
		szDir = ""
		flag = SWShowNormal
	} else if len(args) == 1 {
		if szDir, ok = args[0].(string); !ok {
			panic("szDir must be a string")
		}
		flag = SWShowNormal
	} else if len(args) == 2 {
		if szDir, ok = args[0].(string); !ok {
			panic("szDir must be a string")
		}
		if flag, ok = args[1].(int); !ok {
			panic("flag must be a int")
		}
	} else {
		panic("Too more parameter")
	}
	pid, _, lastErr := runAsWait.Call(strPtr(user), strPtr(domain), strPtr(password), intPtr(loginFlag), strPtr(szProgram), strPtr(szDir), intPtr(flag))
	// log.Println(pid)
	if int(pid) == 0 {
		log.Println(lastErr)
	}
	return int(pid)
}
