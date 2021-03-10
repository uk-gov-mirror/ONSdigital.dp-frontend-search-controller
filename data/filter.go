package data

import (
	"context"
	"net/url"
	"strings"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/log.go/log"
)

// Filter represents information of filters selected by user
type Filter struct {
	Query           []string `json:"query,omitempty"`
	LocaliseKeyName []string `json:"localise_key,omitempty"`
}

// Category represents all the search categories in search page
type Category struct {
	LocaliseKeyName string        `json:"localise_key"`
	Count           int           `json:"count"`
	ContentTypes    []ContentType `json:"content_types"`
}

// ContentType represents the type of the search results and the number of results for each type
type ContentType struct {
	LocaliseKeyName string   `json:"localise_key"`
	Count           int      `json:"count"`
	Type            string   `json:"type"`
	SubTypes        []string `json:"sub_types"`
}

// Categories represent the list of all search categories
var Categories = []Category{Publication, Data, Other}

// Publication - search information on publication category
var Publication = Category{
	LocaliseKeyName: "Publication",
	ContentTypes:    []ContentType{Bulletin, Article, Compendium},
}

// Data - search information on data category
var Data = Category{
	LocaliseKeyName: "Data",
	ContentTypes:    []ContentType{TimeSeries, Datasets, UserRequestedData},
}

// Other - search information on other categories
var Other = Category{
	LocaliseKeyName: "Other",
	ContentTypes:    []ContentType{Methodology, CorporateInformation},
}

// Bulletin - Search information specific for statistical bulletins
var Bulletin = ContentType{
	LocaliseKeyName: "StatisticalBulletin",
	Type:            "bulletin",
	SubTypes:        []string{"bulletin"},
}

// Article - Search information specific for articles
var Article = ContentType{
	LocaliseKeyName: "Article",
	Type:            "article",
	SubTypes:        []string{"article", "article_download"},
}

// Compendium - Search information specific for compendium
var Compendium = ContentType{
	LocaliseKeyName: "Compendium",
	Type:            "compendia",
	SubTypes:        []string{"compendium_landing_page"},
}

// TimeSeries - Search information specific for time series
var TimeSeries = ContentType{
	LocaliseKeyName: "TimeSeries",
	Type:            "time_series",
	SubTypes:        []string{"timeseries"},
}

// Datasets - Search information specific for datasets
var Datasets = ContentType{
	LocaliseKeyName: "Datasets",
	Type:            "datasets",
	SubTypes:        []string{"dataset_landing_page", "reference_tables"},
}

// UserRequestedData - Search information specific for user requested data
var UserRequestedData = ContentType{
	LocaliseKeyName: "UserRequestedData",
	Type:            "user_requested_data",
	SubTypes:        []string{"static_adhoc"},
}

// Methodology - Search information specific for methodologies
var Methodology = ContentType{
	LocaliseKeyName: "Methodology",
	Type:            "methodology",
	SubTypes:        []string{"static_methodology", "static_methodology_download", "static_qmi"},
}

// CorporateInformation - Search information specific for corporate information
var CorporateInformation = ContentType{
	LocaliseKeyName: "CorporateInformation",
	Type:            "corporate_information",
	SubTypes:        []string{"static_foi", "static_page", "static_landing_page", "static_article"},
}

// filterOptions contains all the possible filter available on the search page
var filterOptions = map[string]ContentType{
	Article.Type:              Article,
	Bulletin.Type:             Bulletin,
	Compendium.Type:           Compendium,
	CorporateInformation.Type: CorporateInformation,
	Datasets.Type:             Datasets,
	Methodology.Type:          Methodology,
	TimeSeries.Type:           TimeSeries,
	UserRequestedData.Type:    UserRequestedData,
}

// reviewFilter retrieves filters from query, checks if they are one of the filter options, and updates validatedQueryParams
func reviewFilters(ctx context.Context, urlQuery url.Values, validatedQueryParams *SearchURLParams) error {
	filtersQuery := urlQuery["filter"]

	for _, filterQuery := range filtersQuery {
		filter, found := filterOptions[filterQuery]

		if found {
			validatedQueryParams.Filter.Query = append(validatedQueryParams.Filter.Query, filter.Type)
			validatedQueryParams.Filter.LocaliseKeyName = append(validatedQueryParams.Filter.LocaliseKeyName, filter.LocaliseKeyName)
		} else {
			err := errs.ErrFilterNotFound
			logData := log.Data{"filter not found": filter}
			log.Event(ctx, "failed to find filter", log.Error(err), log.ERROR, logData)

			return err
		}

	}

	return nil
}

// GetCategories returns all the categories and its content types where all the count is set to zero
func GetCategories() []Category {
	var categories []Category
	categories = append(categories, Categories...)

	// To get a different reference of ContentType - deep copy
	for i, category := range categories {
		categories[i].ContentTypes = []ContentType{}
		categories[i].ContentTypes = append(categories[i].ContentTypes, Categories[i].ContentTypes...)

		// To get a different reference of SubTypes - deep copy
		for j := range category.ContentTypes {
			categories[i].ContentTypes[j].SubTypes = []string{}
			categories[i].ContentTypes[j].SubTypes = append(categories[i].ContentTypes[j].SubTypes, Categories[i].ContentTypes[j].SubTypes...)
		}
	}

	return categories
}

// updateQueryWithAPIFilters retrieves and adds all available sub filters which is related to the search filter given by the user
func updateQueryWithAPIFilters(ctx context.Context, apiQuery url.Values) {
	filters := apiQuery["content_type"]

	if len(filters) > 0 {
		subFilters := getSubFilters(ctx, filters)

		apiQuery.Set("content_type", strings.Join(subFilters, ","))
	}
}

// getSubFilters gets all available sub filters which is related to the search filter given by the user
func getSubFilters(ctx context.Context, filters []string) []string {
	var subFilters = make([]string, 0)

	for _, filter := range filters {
		subFilter := filterOptions[filter]
		subFilters = append(subFilters, subFilter.SubTypes...)
	}

	return subFilters
}
