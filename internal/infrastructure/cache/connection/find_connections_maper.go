package cache

import (
	"github.com/d3code/uuid"
	"github.com/nazarkurii/marshrutka_api/internal/entity"
	connectionpb "github.com/nazarkurii/marshrutka_api/internal/proto/proto/connection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProto(r entity.FindConnectionsResponse) *connectionpb.FindConnectionsResponse {
	return &connectionpb.FindConnectionsResponse{
		Connections: mapFoundConnections(r.Connections),
		LeftRange:   mapRanges(r.LeftRange),
		RightRange:  mapRanges(r.RightRange),
	}
}

func fromProto(p *connectionpb.FindConnectionsResponse) entity.FindConnectionsResponse {
	return entity.FindConnectionsResponse{
		Connections: mapFoundConnectionsFromProto(p.Connections),
		LeftRange:   mapRangesFromProto(p.LeftRange),
		RightRange:  mapRangesFromProto(p.RightRange),
	}
}

func mapFoundConnections(
	src []entity.FoundConnection,
) []*connectionpb.FoundConnection {

	if src == nil {
		return nil
	}

	out := make([]*connectionpb.FoundConnection, 0, len(src))
	for _, c := range src {
		out = append(out, &connectionpb.FoundConnection{
			Id:                 c.ID.String(),
			Price:              int32(c.Price),
			Line:               int32(c.Line),
			DepartureCountry:   c.DepartureCountry,
			DestinationCountry: c.DestinationCountry,
			DepartureTime:      timestamppb.New(c.DepartureTime),
			ArrivalTime:        timestamppb.New(c.ArrivalTime),
			EstimatedDuration:  int32(c.EstimatedDuration),
			SellBefore:         timestamppb.New(c.SellBefore),
			TicketsLeft:        int32(c.TicketsLeft),
			Fits:               c.Fits,
		})
	}

	return out
}

func mapFoundConnectionsFromProto(
	src []*connectionpb.FoundConnection,
) []entity.FoundConnection {

	if src == nil {
		return nil
	}

	out := make([]entity.FoundConnection, 0, len(src))
	for _, c := range src {

		id, _ := uuid.Parse(c.Id)

		out = append(out, entity.FoundConnection{
			ConnectionSimplified: entity.ConnectionSimplified{
				ID:                 id,
				Price:              int(c.Price),
				Line:               int(c.Line),
				DepartureCountry:   c.DepartureCountry,
				DestinationCountry: c.DestinationCountry,
				DepartureTime:      c.DepartureTime.AsTime(),
				ArrivalTime:        c.ArrivalTime.AsTime(),
				EstimatedDuration:  int(c.EstimatedDuration),
				SellBefore:         c.SellBefore.AsTime(),
			},
			TicketsLeft: int(c.TicketsLeft),
			Fits:        c.Fits,
		})
	}

	return out
}

func mapRanges(
	src []entity.ConnectionsRange,
) []*connectionpb.ConnectionsRange {

	if src == nil {
		return nil
	}

	out := make([]*connectionpb.ConnectionsRange, 0, len(src))
	for _, r := range src {
		out = append(out, &connectionpb.ConnectionsRange{
			Date:     timestamppb.New(r.Date),
			Number:   int32(r.Number),
			MinPrice: int32(r.MinPrice),
		})
	}

	return out
}

func mapRangesFromProto(
	src []*connectionpb.ConnectionsRange,
) []entity.ConnectionsRange {

	if src == nil {
		return nil
	}

	out := make([]entity.ConnectionsRange, 0, len(src))
	for _, r := range src {
		out = append(out, entity.ConnectionsRange{
			Date:     r.Date.AsTime(),
			Number:   int(r.Number),
			MinPrice: int(r.MinPrice),
		})
	}

	return out
}
