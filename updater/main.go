package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

// Data structures matching your serviceVersions.json
type Entry struct {
	URL      string `json:"url"`
	Snapshot *bool  `json:"snapshot,omitempty"`
}
type VersionMap map[string]Entry

type ServiceVersions struct {
	Proxy struct {
		Velocity   VersionMap `json:"VELOCITY"`
		BungeeCord VersionMap `json:"BUNGEECORD"`
		Waterfall  VersionMap `json:"WATERFALL"`
	} `json:"PROXY"`
	Server struct {
		Paper      VersionMap `json:"PAPER"`
		Pufferfish VersionMap `json:"PUFFERFISH"`
		Purpur     VersionMap `json:"PURPUR"`
		Folia      VersionMap `json:"FOLIA"`
		Vanilla    VersionMap `json:"VANILLA"`
	} `json:"SERVER"`
}

// Get all available versions from API for a project (flattened)
func getPaperLikeVersions(project string) ([]string, error) {
	api := fmt.Sprintf("https://fill.papermc.io/v3/projects/%s", project)
	var result struct {
		Versions map[string][]string `json:"versions"`
	}
	resp, err := http.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	var versions []string
	for _, sub := range result.Versions {
		versions = append(versions, sub...)
	}
	// Sort descending (latest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})
	return versions, nil
}

// Purpur: Get all available versions from API
func getPurpurVersions() ([]string, error) {
	api := "https://api.purpurmc.org/v2/purpur"
	var result struct {
		Versions []string `json:"versions"`
	}
	resp, err := http.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	// Sort descending (latest first)
	sort.Slice(result.Versions, func(i, j int) bool {
		return result.Versions[i] > result.Versions[j]
	})
	return result.Versions, nil
}

// Get latest build download URL for a project/version (new API)
func getProjectLatestDownloadURL(project, version string) (string, error) {
	api := fmt.Sprintf("https://fill.papermc.io/v3/projects/%s/versions/%s/builds/latest", project, version)
	var result struct {
		Downloads map[string]struct {
			URL string `json:"url"`
		} `json:"downloads"`
	}
	resp, err := http.Get(api)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	download, ok := result.Downloads["server:default"]
	if !ok || download.URL == "" {
		return "", fmt.Errorf("no server:default download found for %s %s", project, version)
	}
	return download.URL, nil
}

// Universal updater for a PaperMC project
func updatePaperMCProject(project string, versionMap VersionMap) {
	fmt.Printf("== Checking %s ==\n", strings.Title(project))
	versions, err := getPaperLikeVersions(project)
	if err != nil {
		fmt.Printf("%s: Error loading versions: %v\n", strings.Title(project), err)
		return
	}
	for _, v := range versions {
		key := versionToKey(v)
		url, err := getProjectLatestDownloadURL(project, v)
		if err != nil {
			fmt.Printf("%s %s: Error: %v\n", strings.Title(project), v, err)
			continue
		}
		if entry, ok := versionMap[key]; !ok || entry.URL != url {
			if ok {
				fmt.Printf("%s %s: Updated download URL.\n", strings.Title(project), v)
			} else {
				fmt.Printf("%s %s: Added missing version.\n", strings.Title(project), v)
			}
			versionMap[key] = Entry{URL: url}
		} else {
			fmt.Printf("%s %s: Already up to date.\n", strings.Title(project), v)
		}
	}
}

// Purpur: Get latest build download URL for a version
func getPurpurLatestBuildURL(version string) (string, error) {
	api := fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s", version)
	var result struct {
		Builds struct {
			Latest string `json:"latest"`
		} `json:"builds"`
	}
	resp, err := http.Get(api)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Builds.Latest == "" {
		return "", fmt.Errorf("no latest build found for version %s", version)
	}
	return fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/%s/download", version, result.Builds.Latest), nil
}

// Universal updater for Purpur
func updatePurpurProject(versionMap VersionMap) {
	fmt.Println("== Checking Purpur ==")
	versions, err := getPurpurVersions()
	if err != nil {
		fmt.Printf("Purpur: Error loading versions: %v\n", err)
		return
	}
	for _, v := range versions {
		key := versionToKey(v)
		url, err := getPurpurLatestBuildURL(v)
		if err != nil {
			fmt.Printf("Purpur %s: Error: %v\n", v, err)
			continue
		}
		if entry, ok := versionMap[key]; !ok || entry.URL != url {
			if ok {
				fmt.Printf("Purpur %s: Updated download URL.\n", v)
			} else {
				fmt.Printf("Purpur %s: Added missing version.\n", v)
			}
			versionMap[key] = Entry{URL: url}
		} else {
			fmt.Printf("Purpur %s: Already up to date.\n", v)
		}
	}
}

// Liefert alle Vanilla-Vollversionen (neueste zuerst)
func getVanillaReleaseVersions() ([]string, error) {
	api := "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"
	var result struct {
		Versions []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"versions"`
	}
	resp, err := http.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	var releases []string
	for _, v := range result.Versions {
		if v.Type == "release" {
			releases = append(releases, v.ID)
		}
	}
	return releases, nil
}

// Holt die Download-URL für die Server-JAR einer bestimmten Vanilla-Version
func getVanillaDownloadURL(version string) (string, error) {
	manifestURL := "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"
	var manifest struct {
		Versions []struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"versions"`
	}
	resp, err := http.Get(manifestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", err
	}
	var versionURL string
	for _, v := range manifest.Versions {
		if v.ID == version {
			versionURL = v.URL
			break
		}
	}
	if versionURL == "" {
		return "", fmt.Errorf("version %s not found in manifest", version)
	}
	var versionManifest struct {
		Downloads struct {
			Server struct {
				URL string `json:"url"`
			} `json:"server"`
		} `json:"downloads"`
	}
	resp2, err := http.Get(versionURL)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()
	if err := json.NewDecoder(resp2.Body).Decode(&versionManifest); err != nil {
		return "", err
	}
	if versionManifest.Downloads.Server.URL == "" {
		return "", fmt.Errorf("no server download found for version %s", version)
	}
	return versionManifest.Downloads.Server.URL, nil
}

func updateVanillaReleaseProject(versionMap VersionMap) {
	fmt.Println("== Checking Vanilla (Releases only) ==")
	versions, err := getVanillaReleaseVersions()
	if err != nil {
		fmt.Printf("Vanilla: Error loading versions: %v\n", err)
		return
	}
	for _, v := range versions {
		key := versionToKey(v)
		url, err := getVanillaDownloadURL(v)
		if err != nil {
			fmt.Printf("Vanilla %s: Error: %v\n", v, err)
			continue
		}
		if entry, ok := versionMap[key]; !ok || entry.URL != url {
			if ok {
				fmt.Printf("Vanilla %s: Updated download URL.\n", v)
			} else {
				fmt.Printf("Vanilla %s: Added missing version.\n", v)
			}
			versionMap[key] = Entry{URL: url}
		} else {
			fmt.Printf("Vanilla %s: Already up to date.\n", v)
		}
	}
}

// For pretty JSON keys (replace . with _)
func versionToKey(version string) string {
	return strings.ReplaceAll(version, ".", "_")
}

func main() {
	// Load file
	file, err := os.Open("serviceVersions.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var data ServiceVersions
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		panic(err)
	}

	// Update all projects
	updatePaperMCProject("paper", data.Server.Paper)
	updatePaperMCProject("folia", data.Server.Folia)
	updatePaperMCProject("velocity", data.Proxy.Velocity)
	updatePaperMCProject("waterfall", data.Proxy.Waterfall)
	updatePurpurProject(data.Server.Purpur)
	updateVanillaReleaseProject(data.Server.Vanilla)

	// Write file back
	out, err := os.Create("serviceVersions.json")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		panic(err)
	}
	fmt.Println("serviceVersions.json has been updated!")
}
