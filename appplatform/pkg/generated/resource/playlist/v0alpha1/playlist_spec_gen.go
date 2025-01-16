package v0alpha1

// Defines values for ItemType.
const (
	ItemTypeDashboardById  ItemType = "dashboard_by_id"
	ItemTypeDashboardByTag ItemType = "dashboard_by_tag"
	ItemTypeDashboardByUid ItemType = "dashboard_by_uid"
)

// Item defines model for Item.
// +k8s:openapi-gen=true
type Item struct {
	// type of the item.
	Type ItemType `json:"type"`

	// Value depends on type and describes the playlist item.
	//  - dashboard_by_id: The value is an internal numerical identifier set by Grafana. This
	//  is not portable as the numerical identifier is non-deterministic between different instances.
	//  Will be replaced by dashboard_by_uid in the future. (deprecated)
	//  - dashboard_by_tag: The value is a tag which is set on any number of dashboards. All
	//  dashboards behind the tag will be added to the playlist.
	//  - dashboard_by_uid: The value is the dashboard UID
	Value string `json:"value"`
}

// ItemType type of the item.
// +k8s:openapi-gen=true
type ItemType string

// Spec defines model for Spec.
// +k8s:openapi-gen=true
type Spec struct {
	Interval string `json:"interval"`
	Items    []Item `json:"items"`
	Title    string `json:"title"`
}
