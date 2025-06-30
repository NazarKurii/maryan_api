package trip

import (
	"fmt"
	"maryan_api/config"
	"maryan_api/pkg/hypermedia"
	"maryan_api/pkg/images"
	rfc7807 "maryan_api/pkg/problem"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type busService interface {
	create(bus Bus, busImages []*multipart.FileHeader) (uuid.UUID, error)
	get(id string) (Bus, error)
	getBuses(page, size string) ([]Bus, hypermedia.Links, error)
	delete(id string) error
	makeActive(id string) error
	makeInActive(id string) error
}

type busServiceImpl struct {
	repo busRepo
}

func (bs *busServiceImpl) create(bus Bus, busImages []*multipart.FileHeader) (uuid.UUID, error) {
	params := bus.prepare()
	if len(busImages) == 0 {
		params.SetInvalidParam("images", "No images atached")
	}
	if len(params) != 0 {
		return uuid.Nil, rfc7807.BadRequest("invalid-bus-data", "Invalid Bus Data Error", "Invalid params.", params...)
	}

	for i, image := range busImages {
		id := uuid.NewString()
		if err := images.Save("../../../static/imgs/"+id, image); err != nil {
			params.SetInvalidParam(fmt.Sprintf("image(index:%d)", i), err.Error())
		} else {
			bus.Images = append(bus.Images, BusImage{bus.ID, config.APIURL() + "/imgs/" + id})
		}
	}

	if len(params) != 0 {
		return uuid.Nil, rfc7807.BadRequest("invalid-bus-data", "Invalid Bus Data Error", "Invalid params.", params...)
	}

	return bus.ID, bs.repo.create(&bus)
}

func (bs *busServiceImpl) get(id string) (Bus, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return Bus{}, rfc7807.UUID(err.Error())
	}
	return bs.repo.get(uuid)
}

func (bs *busServiceImpl) getBuses(page, size string) ([]Bus, hypermedia.Links, error) {
	var params rfc7807.InvalidParams

	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		params.SetInvalidParam("pageNumber", err.Error())
	} else if pageNumber < 1 {
		params.SetInvalidParam("pageNumber", "Must be equal or greater than 1.")
	}

	pageSize, err := strconv.Atoi(size)
	if err != nil {
		params.SetInvalidParam("pageSize", err.Error())
	} else if pageSize < 1 {
		params.SetInvalidParam("pageSize", "Must be equal or greater than 1.")
	}

	if params != nil {
		return nil, nil, rfc7807.BadRequest("invalid-page-data", "Page Data Error", "invalid request params.", params...)
	}

	buses, pages, err := bs.repo.getBuses(pageNumber, pageSize)
	if err != nil {
		return nil, nil, err
	}

	var pagesUrls = make(hypermedia.Links, pages)
	for i := 0; i < pages; i++ {
		pagesUrls[i] = hypermedia.Link{strconv.Itoa(i + 1): hypermedia.Href{config.APIURL() + fmt.Sprintf("/admin/buses/%d/%d", i+1, pageSize), http.MethodGet}}
	}

	return buses, pagesUrls, nil
}

func (bs *busServiceImpl) delete(id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return rfc7807.UUID(err.Error())
	}
	return bs.repo.delete(uuid)

}

func (bs *busServiceImpl) makeActive(id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return rfc7807.UUID(err.Error())
	}

	isActive, err := bs.repo.isActive(uuid)
	if err != nil {
		return err
	}

	if isActive {
		return rfc7807.BadRequest("already-active", "Already Active Bus Error", "The bus already has an 'Active' status.")
	}

	return bs.repo.makeActive(uuid)
}

func (bs *busServiceImpl) makeInActive(id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return rfc7807.UUID(err.Error())
	}

	isActive, err := bs.repo.isActive(uuid)
	if err != nil {
		return err
	}

	if !isActive {
		return rfc7807.BadRequest("already-inactive", "Already Inactive Bus Error", "The bus already has an 'Inactive' status.")
	}

	return bs.repo.makeInactive(uuid)
}

type passengerService interface {
}

type passengerServiceImpl struct {
	repo   passengerRepo
	client *http.Client
}

// --------------------Servises Initialization Functions
func newBusService(bus busRepo) busService {
	return &busServiceImpl{bus}
}

func newPassengerService(passenger passengerRepo, client *http.Client) passengerService {
	return &passengerServiceImpl{passenger, client}
}
