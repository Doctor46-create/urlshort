package server

import (
	"net"
	"strings"
	"context"
	"google.golang.org/grpc/metadata"
)

func IsIPInSubnet(ipStr, subnetCIDR string) bool {
	if subnetCIDR == "" {
		return false
	}

	ip := net.ParseIP(ipStr)
	_, subnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return false
	}

	return subnet.Contains(ip)
}

func ExtractIP(ipHeader string) string {
	return strings.TrimSpace(ipHeader)
}

func getRequestID(ctx context.Context) string {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return ""
    }

    ids := md.Get("x-request-id") 
    if len(ids) > 0 {
        return ids[0]
    }

    return ""
}

