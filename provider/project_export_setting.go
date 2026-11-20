package provider

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const projectExportSettingField = "export_setting"

func getProjectExportSettingSchema(computed bool) *schema.Schema {
	urlSchema := &schema.Schema{
		Type:        schema.TypeString,
		Required:    !computed,
		Computed:    computed,
		Description: "The HTTPS SharePoint Online URL of a document library or folder. Read values use the canonical URL.",
	}
	description := "The project export destination."
	if !computed {
		description += " Omit the block to preserve the current destination; configure an empty `export_setting {}` block to clear it."
		urlSchema.ValidateFunc = func(value interface{}, _ string) ([]string, []error) {
			if _, err := normalizeProjectSharePointURL(value.(string)); err != nil {
				return nil, []error{err}
			}
			return nil, nil
		}
		urlSchema.StateFunc = func(value interface{}) string {
			raw := value.(string)
			canonical, err := normalizeProjectSharePointURL(raw)
			if err != nil {
				return raw
			}
			return canonical
		}
	}

	sharepoint := &schema.Schema{
		Type:        schema.TypeList,
		Optional:    !computed,
		Computed:    computed,
		Description: "The project's SharePoint export destination.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{"url": urlSchema},
		},
	}
	exportSetting := &schema.Schema{
		Type:        schema.TypeList,
		Optional:    !computed,
		Computed:    true,
		Description: description,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{"sharepoint": sharepoint},
		},
	}
	if !computed {
		sharepoint.MaxItems = 1
		exportSetting.MaxItems = 1
	}
	return exportSetting
}

// Read the raw configuration rather than state to distinguish omission from
// an explicit empty block, which asks Bytebase to clear the destination.
func projectExportSettingFromConfig(raw cty.Value) (*v1pb.Project_ExportSetting, bool, error) {
	if raw.IsNull() || !raw.IsKnown() {
		return nil, false, nil
	}
	config := raw.GetAttr(projectExportSettingField)
	if config.IsNull() || (config.IsKnown() && config.LengthInt() == 0) {
		return nil, false, nil
	}
	if !config.IsWhollyKnown() {
		return nil, false, errors.New("the SharePoint export destination is not known")
	}
	sharepoint := config.Index(cty.NumberIntVal(0)).GetAttr("sharepoint")
	if sharepoint.IsNull() || sharepoint.LengthInt() == 0 {
		return nil, true, nil
	}
	rawURL := sharepoint.Index(cty.NumberIntVal(0)).GetAttr("url")
	if rawURL.IsNull() {
		return nil, false, errors.New("the SharePoint URL must be configured")
	}
	canonical, err := normalizeProjectSharePointURL(rawURL.AsString())
	if err != nil {
		return nil, false, err
	}
	return &v1pb.Project_ExportSetting{
		Sharepoint: &v1pb.SharePointDestination{Url: canonical},
	}, true, nil
}

func flattenProjectExportSetting(setting *v1pb.Project_ExportSetting, resourceRead bool) []interface{} {
	if url := setting.GetSharepoint().GetUrl(); url != "" {
		return []interface{}{map[string]interface{}{
			"sharepoint": []interface{}{map[string]interface{}{"url": url}},
		}}
	}
	if resourceRead {
		// Canonicalize the cleared state to the same shape as export_setting {}.
		return []interface{}{map[string]interface{}{"sharepoint": []interface{}{}}}
	}
	return nil
}

var (
	sharePointCopiedLinkPattern  = regexp.MustCompile(`^/:[A-Za-z]:/([^/]*)(/.*)?$`)
	sharePointLibraryViewPattern = regexp.MustCompile(`(?i)/Forms/[^/]+\.aspx$`)
)

// Mirrors the Bytebase server's pure normalization contract
// (backend/plugin/sharepoint/url.go). Resolution and access checks are server-owned.
func normalizeProjectSharePointURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return "", errors.New("enter the full SharePoint URL of a document library or folder, starting with https://")
	}
	if u.Scheme != "https" {
		return "", errors.New("the SharePoint URL must start with https://")
	}
	if u.User != nil || u.Port() != "" {
		return "", errors.New("the SharePoint URL must not contain credentials or a port")
	}
	host := strings.ToLower(u.Hostname())
	tenant, ok := strings.CutSuffix(host, ".sharepoint.com")
	if !ok || tenant == "" || strings.Contains(tenant, ".") {
		return "", errors.New("the SharePoint URL must be on a *.sharepoint.com host")
	}
	if strings.HasSuffix(tenant, "-my") {
		return "", errors.New("OneDrive URLs are not supported; use a document library or folder on a SharePoint site")
	}

	p := u.Path
	if m := sharePointCopiedLinkPattern.FindStringSubmatch(p); m != nil {
		if !strings.EqualFold(m[1], "r") {
			return "", errors.New("SharePoint sharing links are not supported; use the folder URL from the browser's address bar")
		}
		p = m[2]
	} else if strings.HasSuffix(strings.ToLower(p), ".aspx") {
		query := u.Query()
		if id := query.Get("id"); id != "" {
			p = id
		} else if rootFolder := query.Get("RootFolder"); rootFolder != "" {
			p = rootFolder
		} else {
			p = sharePointLibraryViewPattern.ReplaceAllString(p, "")
		}
	}

	var segments []string
	for _, segment := range strings.Split(p, "/") {
		switch {
		case segment == "":
		case segment == "." || segment == "..":
			return "", errors.New("the SharePoint URL must not contain . or .. path segments")
		case strings.EqualFold(segment, "_layouts") || strings.HasSuffix(strings.ToLower(segment), ".aspx"):
			return "", errors.New("the SharePoint URL is a page, not a document library or folder")
		default:
			segments = append(segments, segment)
		}
	}
	if len(segments) == 0 {
		return "", errors.New("the SharePoint URL points to a site, not a document library or folder")
	}
	if first := strings.ToLower(segments[0]); (first == "sites" || first == "teams") && len(segments) <= 2 {
		return "", errors.New("the SharePoint URL points to a site, not a document library or folder")
	}
	canonical := url.URL{Scheme: "https", Host: host, Path: "/" + strings.Join(segments, "/")}
	return canonical.String(), nil
}
