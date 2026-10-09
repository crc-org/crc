/*
 * This file is part of the libvirt-go-xml-module project
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * Copyright (C) 2016 Red Hat, Inc.
 *
 */

package libvirtxml

import (
	"encoding/xml"
	"fmt"
)

type SysInfoBIOS struct {
	Entry []SysInfoEntry `xml:"entry"`
}

type SysInfoSystem struct {
	Entry []SysInfoEntry `xml:"entry"`
}

type SysInfoBaseBoard struct {
	Entry []SysInfoEntry `xml:"entry"`
}

type SysInfoProcessor struct {
	Entry []SysInfoEntry `xml:"entry"`
}

type SysInfoMemory struct {
	Entry []SysInfoEntry `xml:"entry"`
}

type SysInfoChassis struct {
	Entry []SysInfoEntry `xml:"entry"`
}

type SysInfoOEMStrings struct {
	Entry []string `xml:"entry"`
}

type SysInfoEntry struct {
	Name  string `xml:"name,attr"`
	File  string `xml:"file,attr,omitempty"`
	Value string `xml:",chardata"`
}

type SysInfoSMBIOS struct {
	BIOS       *SysInfoBIOS       `xml:"bios"`
	System     *SysInfoSystem     `xml:"system"`
	BaseBoard  []SysInfoBaseBoard `xml:"baseBoard"`
	Chassis    *SysInfoChassis    `xml:"chassis"`
	Processor  []SysInfoProcessor `xml:"processor"`
	Memory     []SysInfoMemory    `xml:"memory_device"`
	OEMStrings *SysInfoOEMStrings `xml:"oemStrings"`
}

type SysInfoFWCfg struct {
	Entry []SysInfoEntry `xml:"entry"`
}

type SysInfo struct {
	SMBIOS *SysInfoSMBIOS `xml:"-"`
	FWCfg  *SysInfoFWCfg  `xml:"-"`
}

func (d *SysInfo) Unmarshal(doc string) error {
	return xml.Unmarshal([]byte(doc), d)
}

func (d *SysInfo) Marshal() (string, error) {
	doc, err := xml.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(doc), nil
}

type sysInfo SysInfo

type sysInfoSMBIOS struct {
	SysInfoSMBIOS
	sysInfo
}

type sysInfoFWCfg struct {
	SysInfoFWCfg
	sysInfo
}

func (a *SysInfo) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "sysinfo"
	if a.SMBIOS != nil {
		smbios := sysInfoSMBIOS{}
		smbios.sysInfo = sysInfo(*a)
		smbios.SysInfoSMBIOS = *a.SMBIOS
		start.Attr = append(start.Attr, xml.Attr{
			xml.Name{Local: "type"}, "smbios",
		})
		return e.EncodeElement(smbios, start)
	} else if a.FWCfg != nil {
		fwcfg := sysInfoFWCfg{}
		fwcfg.sysInfo = sysInfo(*a)
		fwcfg.SysInfoFWCfg = *a.FWCfg
		start.Attr = append(start.Attr, xml.Attr{
			xml.Name{Local: "type"}, "fwcfg",
		})
		return e.EncodeElement(fwcfg, start)
	} else {
		gen := sysInfo(*a)
		return e.EncodeElement(gen, start)
	}
}

func (a *SysInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	typ, ok := getAttr(start.Attr, "type")
	if !ok {
		return fmt.Errorf("Missing 'type' attribute on  controller")
	}
	if typ == "smbios" {
		var smbios sysInfoSMBIOS
		err := d.DecodeElement(&smbios, &start)
		if err != nil {
			return err
		}
		*a = SysInfo(smbios.sysInfo)
		a.SMBIOS = &smbios.SysInfoSMBIOS
		return nil
	} else if typ == "fwcfg" {
		var fwcfg sysInfoFWCfg
		err := d.DecodeElement(&fwcfg, &start)
		if err != nil {
			return err
		}
		*a = SysInfo(fwcfg.sysInfo)
		a.FWCfg = &fwcfg.SysInfoFWCfg
		return nil
	} else {
		var gen sysInfo
		err := d.DecodeElement(&gen, &start)
		if err != nil {
			return err
		}
		*a = SysInfo(gen)
		return nil
	}
}
