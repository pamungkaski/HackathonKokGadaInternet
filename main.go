package main

import (
	"fmt"
	"github.com/Genwis/Genwis-Saiyan-Mode/helpers"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"os"
	"io/ioutil"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"io"
	"crypto/rand"
)

type Itinerary struct {
	ID       string  `json:"id" gorm:"primary_key"`
	Cost     float64 `json:"cost"`
	Username string  `json:"username"`
	Detail   Detail  `json:"detail"`
	TimeLine []Daily `json:"time_line"`
}

type Location struct {
	ID        string     `json:"id"`
	City      string     `json:"city"`
	State     string     `json:"state"`
	Country   string     `json:"country"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type Detail struct {
	LocationID string    `json:"location_id"`
	Location   Location  `gorm:"association_foreignkey:ID" json:"location"`
	Start      time.Time `json:"start"`
	Finish     time.Time `json:"finish"`
	Budget     float64   `json:"budget"`
	Tags       Tags      `json:"tags"`
	Pax        Pax       `json:"pax"`
}

type Tags struct {
	Culture        bool `json:"culture"`
	Outdoors       bool `json:"outdoors"`
	History        bool `json:"history"`
	Shopping       bool `json:"shopping"`
	Wildlife       bool `json:"wildlife"`
	Beaches        bool `json:"beaches"`
	Mountain       bool `json:"mountain"`
	Museum         bool `json:"museum"`
	Amusement      bool `json:"amusement"`
	HiddenParadise bool `json:"hidden_paradise"`
}

type Daily struct {
	Time     time.Time `json:"time"`
	Events   []Event   `json:"events"`
	Traffics []Edge    `json:"traffic"`
	End      time.Time `json:"end"`
}

type Event struct {
	Start        time.Time  `json:"start"`
	End          time.Time  `json:"end"`
	Attraction   Attraction `json:"attraction" gorm:"foreignkey:AttractionID"`
	AttractionID string     `json:"attraction_id"`
}

type Pax struct {
	Adult  int `json:"adult"`
	Child  int `json:"child"`
	Infant int `json:"infant"`
}

type Edge struct {
	OriginID      string `json:"origin_id"`
	DestinationID string `json:"destination_id"`
	Distance      int    `json:"distance"`
	TravelTime    int    `json:"travel_time"`
}

type DistanceMatrixResponse struct {
	DestinationAddresses []string                 `json:"destination_addresses"`
	OriginAddresses      []string                 `json:"origin_addresses"`
	Rows                 []DistanceMatrixElements `json:"rows"`
	Status               string                   `json:"status"`
}

type DistanceMatrixElements struct {
	Elements []DistanceMatrixElement `json:"elements"`
}

type DistanceMatrixElement struct {
	Distance DistanceMatrixDistance `json:"distance"`
	Duration DistanceMatrixDuration `json:"duration"`
	Status   string                 `json:"status"`
}

type DistanceMatrixDistance struct {
	Text  string `json:"text"`
	Value int    `json:"value"`
}

type DistanceMatrixDuration struct {
	Text  string `json:"text"`
	Value int    `json:"value"`
}

type ItineraryRequest struct {
	LocationID string   `json:"location_id"`
	Location   Location `gorm:"association_foreignkey:ID"`
	Start      string   `json:"start"`
	Finish     string   `json:"finish"`
	Budget     float64  `json:"budget"`
	Tags       Tags     `json:"tags"`
	Pax        Pax      `json:"pax"`
}

type AttractionSimilarly struct {
	Similarity float64
	Attraction Attraction
}
type keyTraf struct {
	Key      int
	Traffics Edge
}

type Attraction struct {
	ID                  string     `json:"id" gorm:"primary_key"`
	Name                string     `json:"name"`
	Description         string     `json:"description"`
	Website             string     `json:"website"`
	PhoneNumber         string     `json:"phone_number"`
	RecommendedDuration int        `json:"recommended_duration"`
	Vicinity            string     `json:"vicinity"`
	Price               float64    `json:"price"`
	Rating              float64    `json:"rating"`
	PartnerUsername     string     `json:"partner_username"`
	Route               string     `json:"route"`
	LocationID          string     `json:"location_id"`
	Location            Location   `json:"location" gorm:"association_foreignkey:ID"`
	Coordinate          Coordinate `json:"coordinate"`
	Tags                Tags       `json:"tags"`
	Photo               []string   `json:"photo"`
	OpeningHours        map[string]OpClose  `json:"opening_hours"`
}

type Coordinate struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// Open/Close Structure
type OpClose struct {
	Open bool   `json:"open"`
	Time CPTime `json:"time"`
}

// Close Open Time Structure
type CPTime struct {
	Close int `json:"close"`
	Open  int `json:"open"`
}

func GetAllAttraction() ([]Attraction, error) {
	var data map[string]map[string]map[string]Attraction
	var list []Attraction
	// Open our jsonFile
	jsonFile, err := os.Open("database.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		return list, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &data)

	for _, k := range data["attractions"]["8ec9ee93-8863-419a-96f9-9a2a4cc7d815"] {
		list = append(list, k)
	}

	return list, nil
}

func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func CalculateDistanceToManyCoordinate(dest Attraction, aorigin []Attraction, key string) ([]Edge, error) {
	var edges []Edge
	desLat, err := strconv.ParseFloat(strings.Replace(dest.Coordinate.Latitude, " ", "", -1), 64)
	if err != nil {
		return nil, err
	}
	desLong, err := strconv.ParseFloat(strings.Replace(dest.Coordinate.Longitude, " ", "", -1), 64)
	if err != nil {
		return nil, err
	}
	for _, tujuan := range aorigin {
		R := 6371.0
		lat, err := strconv.ParseFloat(strings.Replace(tujuan.Coordinate.Latitude, " ", "", -1), 64)
		if err != nil {
			return nil, err
		}
		long, err := strconv.ParseFloat(strings.Replace(tujuan.Coordinate.Longitude, " ", "", -1), 64)
		if err != nil {
			return nil, err
		}
		dLat := deg2rad(lat - desLat)
		dLong := deg2rad(long - desLong)

		a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(deg2rad(desLat))*math.Cos(deg2rad(lat))*math.Pow(math.Sin(dLong/2), 2)

		c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
		d := R * c
		ed := Edge{
			Distance:   int(d * 1000),
			TravelTime: int(d * 80),
		}
		edges = append(edges, ed)
	}

	return edges, nil
}

func deg2rad(deg float64) float64 {
	return math.Phi * deg / 180
}

func WeekdayToInt(Weekday time.Weekday) int {
	var day int
	switch Weekday {
	case time.Monday:
		day = 0
	case time.Tuesday:
		day = 1
	case time.Wednesday:
		day = 2
	case time.Thursday:
		day = 3
	case time.Friday:
		day = 4
	case time.Saturday:
		day = 5
	case time.Sunday:
		day = 6
	}
	return day
}

func CalculateSimilarity(detailTags Tags, attractionDetails Tags) int {
	Similarity := 0
	if detailTags.Culture != attractionDetails.Culture {
		Similarity += 1
	}
	if detailTags.Outdoors != attractionDetails.Outdoors {
		Similarity += 1
	}
	if detailTags.History != attractionDetails.History {
		Similarity += 1
	}
	if detailTags.Shopping != attractionDetails.Shopping {
		Similarity += 1
	}
	if detailTags.Beaches != attractionDetails.Beaches {
		Similarity += 1
	}
	if detailTags.Mountain != attractionDetails.Mountain {
		Similarity += 1
	}
	if detailTags.Museum != attractionDetails.Museum {
		Similarity += 1
	}
	if detailTags.Amusement != attractionDetails.Amusement {
		Similarity += 1
	}
	if detailTags.HiddenParadise != attractionDetails.HiddenParadise {
		Similarity += 1
	}

	return Similarity
}

func CalculateSimilarityList(detail Detail, AttractionList []Attraction) ([]float64, error) {
	var similarity []float64
	for i := 0; i < len(AttractionList); i += 1 {
		similarity = append(similarity, float64(CalculateSimilarity(detail.Tags, AttractionList[i].Tags)))
	}
	return similarity, nil
}

func CreateItinerary(detail Detail, username string) (Itinerary, error) {
	var itinerary Itinerary
	EstimatedCost := 0.0
	itinerary.ID, _ = NewUUID()
	itinerary.Username = username
	itinerary.Detail = detail

	// Get attraction for this location
	AttractionList, err := GetAllAttraction()
	if err != nil {
		return Itinerary{}, err
	}

	if err != nil {
		return Itinerary{}, err
	}
	SimilarityList, err := CalculateSimilarityList(detail, AttractionList)
	if err != nil {
		return Itinerary{}, err
	}

	var AttSim []AttractionSimilarly
	for i, att := range AttractionList {
		AttSim = append(AttSim, AttractionSimilarly{Similarity: SimilarityList[i], Attraction: att})
	}

	sort.Slice(AttSim, func(i, j int) bool {
		return AttSim[i].Similarity < AttSim[j].Similarity
	})
	var sortedAttraction []Attraction
	for _, att := range AttSim {
		sortedAttraction = append(sortedAttraction, att.Attraction)
	}

	AttractionList = sortedAttraction

	Duration := int(detail.Finish.Sub(detail.Start).Hours()/24) + 1

	for i := 0; i < Duration; i += 1 {
		itinerary.TimeLine = append(itinerary.TimeLine, Daily{})
	}

	itinerary.TimeLine[0].Time = time.Date(detail.Start.Year(), detail.Start.Month(), detail.Start.Day(), 0, 0, 0, 0, time.Local)
	itinerary.TimeLine[0].End = time.Date(detail.Start.Year(), detail.Start.Month(), detail.Start.Day(), 8, 0, 0, 0, time.Local)
	for i := 1; i < Duration; i++ {
		itinerary.TimeLine[i].Time = itinerary.TimeLine[i-1].Time.AddDate(0, 0, 1).Local()
		itinerary.TimeLine[i].End = itinerary.TimeLine[i-1].End.AddDate(0, 0, 1).Local()
	}

	AttractionIterator := 0

	Day := 0

	coorlist := make([]string, Duration)
	atrList := make([]Attraction, Duration)

	for EstimatedCost <= detail.Budget && len(AttractionList) > AttractionIterator && Duration > Day {
		Weekday := helpers.WeekdayToInt(itinerary.TimeLine[Day].Time.Weekday())
		projectedEnd := itinerary.TimeLine[Day].End.Add(time.Minute * time.Duration(AttractionList[AttractionIterator].RecommendedDuration))
		closeHour := itinerary.TimeLine[Day].Time.Add(time.Minute * time.Duration(AttractionList[AttractionIterator].OpeningHours[fmt.Sprintf("%d", Weekday)].Time.Close))
		if closeHour.After(projectedEnd) && (EstimatedCost+AttractionList[AttractionIterator].Price) <= detail.Budget && projectedEnd.Day() == itinerary.TimeLine[Day].End.Day() {
			EstimatedCost += AttractionList[AttractionIterator].Price

			Event := Event{
				AttractionID: AttractionList[AttractionIterator].ID,
				Attraction:   AttractionList[AttractionIterator],
				Start:        itinerary.TimeLine[Day].End.Local(),
				End:          projectedEnd.Local(),
			}

			coorlist[Day] = fmt.Sprintf("%v, %v", Event.Attraction.Coordinate.Latitude, Event.Attraction.Coordinate.Longitude)
			atrList[Day] = Event.Attraction
			itinerary.TimeLine[Day].Events = append(itinerary.TimeLine[Day].Events, Event)
			itinerary.TimeLine[Day].End = projectedEnd.Local()
			AttractionIterator += 1
			Day += 1
		} else {
			AttractionIterator += 1
		}
	}

	for EstimatedCost <= detail.Budget && len(AttractionList) > AttractionIterator {
		traffics, err := CalculateDistanceToManyCoordinate(AttractionList[AttractionIterator], atrList, "")
		if err != nil {
			return Itinerary{}, err
		}
		var key []int

		var kt []keyTraf
		for j, t := range traffics {
			kt = append(kt, keyTraf{j, t})
		}

		//fmt.Printf("%v,  %v\n", len(key), len(traffics))
		sort.Slice(kt, func(i, j int) bool {
			return kt[i].Traffics.Distance < kt[j].Traffics.Distance
		})

		for j := range kt {
			key = append(key, j)
		}

		Day = 0
		if traffics[0].Distance < 21000 {
			for Day < Duration {
				Weekday := helpers.WeekdayToInt(itinerary.TimeLine[key[Day]].Time.Weekday())
				projectedEnd := itinerary.TimeLine[key[Day]].End.Add(time.Minute * time.Duration(AttractionList[AttractionIterator].RecommendedDuration))
				projectedEnd = projectedEnd.Add(time.Second * time.Duration(traffics[key[Day]].TravelTime+900))
				closeHour := itinerary.TimeLine[key[Day]].Time.Add(time.Minute * time.Duration(AttractionList[AttractionIterator].OpeningHours[fmt.Sprintf("%d", Weekday)].Time.Close))
				if closeHour.After(projectedEnd) && (EstimatedCost+AttractionList[AttractionIterator].Price) <= detail.Budget && projectedEnd.Day() == itinerary.TimeLine[key[Day]].End.Day() {
					EstimatedCost += AttractionList[AttractionIterator].Price

					traffics[key[Day]].TravelTime += 900

					Event := Event{
						AttractionID: AttractionList[AttractionIterator].ID,
						Attraction:   AttractionList[AttractionIterator],
						Start:        itinerary.TimeLine[key[Day]].End.Add(time.Second * time.Duration(traffics[key[Day]].TravelTime)).Local(),
						End:          projectedEnd.Local(),
					}

					coorlist[key[Day]] = fmt.Sprintf("%v, %v", Event.Attraction.Coordinate.Latitude, Event.Attraction.Coordinate.Longitude)
					atrList[key[Day]] = Event.Attraction

					itinerary.TimeLine[key[Day]].Traffics = append(itinerary.TimeLine[key[Day]].Traffics, traffics[key[Day]])
					itinerary.TimeLine[key[Day]].Events = append(itinerary.TimeLine[key[Day]].Events, Event)
					itinerary.TimeLine[key[Day]].End = projectedEnd.Local()
					// skip
					Day = Duration
				} else {
					Day += 1
				}
			}
		} else {
			for x := AttractionIterator; x < Duration && x < len(AttractionList); x++ {
				tmp := AttractionList[x]
				AttractionList[x] = AttractionList[x+1]
				AttractionList[x+1] = tmp
			}
		}

		AttractionIterator += 1
	}

	itinerary.Cost = EstimatedCost

	return itinerary, nil
}

func main() {
	r := gin.Default()
	r.POST("/itinerary", func(ctx *gin.Context) {
		request := ItineraryRequest{}
		if err := ctx.BindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		Finish, _ := time.Parse("2006-Jan-02", request.Finish)
		Start, _ := time.Parse("2006-Jan-02", request.Start)

		requestData := Detail{
			LocationID: request.LocationID,
			Tags:       request.Tags,
			Budget:     request.Budget,
			Start:      Start,
			Finish:     Finish,
		}

		Itinerary, err := CreateItinerary(requestData, "heheh")

		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, Itinerary)
	})
	r.Run()
}
