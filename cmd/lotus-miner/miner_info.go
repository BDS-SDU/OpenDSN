package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/filecoin-project/lotus/chain/types"
	sealing "github.com/filecoin-project/lotus/storage/pipeline"
)

const minerInfoUpdateInterval = 60 * time.Second

type minerInfoSnapshot struct {
	IP                   string
	Miner                string
	QualityAdjustedPower string
	CommittedSpace       string
	UserDataSize         string
}

func startMinerInfoMonitor(
	ctx context.Context,
	shutdownChan <-chan struct{},
	minerapi api.StorageMiner,
	fullapi v1api.FullNode,
	logPath string,
) {
	update := func() {
		if err := updateMinerInfoLog(ctx, minerapi, fullapi, logPath); err != nil {
			log.Errorf("updating miner info log: %s", err)
		}
	}

	update()

	go func() {
		ticker := time.NewTicker(minerInfoUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				update()
			case <-shutdownChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func updateMinerInfoLog(
	ctx context.Context,
	minerapi api.StorageMiner,
	fullapi v1api.FullNode,
	logPath string,
) error {
	info, err := collectMinerInfoSnapshot(ctx, minerapi, fullapi)
	if err != nil {
		return err
	}

	return writeMinerInfoLog(logPath, info)
}

func collectMinerInfoSnapshot(
	ctx context.Context,
	minerapi api.StorageMiner,
	fullapi v1api.FullNode,
) (*minerInfoSnapshot, error) {
	ip, err := detectHostIP()
	if err != nil {
		ip = "unknown"
	}

	maddr, err := minerapi.ActorAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("get miner actor address: %w", err)
	}

	pow, err := fullapi.StateMinerPower(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("get miner power: %w", err)
	}

	committedSpace, err := computeCommittedSpace(ctx, fullapi, maddr)
	if err != nil {
		return nil, err
	}

	userDataSize, err := computeUserDataSize(ctx, minerapi)
	if err != nil {
		userDataSize = "unknown"
	}

	return &minerInfoSnapshot{
		IP:                   ip,
		Miner:                maddr.String(),
		QualityAdjustedPower: types.DeciStr(pow.MinerPower.QualityAdjPower),
		CommittedSpace:       committedSpace,
		UserDataSize:         userDataSize,
	}, nil
}

func detectHostIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("list network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip4 := ip.To4()
			if ip4 == nil {
				continue
			}

			if !ip4.IsGlobalUnicast() {
				continue
			}

			return ip4.String(), nil
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("list interface addresses: %w", err)
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip4 := ip.To4()
		if ip4 == nil {
			continue
		}
		return ip4.String(), nil
	}

	return "", fmt.Errorf("no non-loopback IPv4 address found")
}

func computeCommittedSpace(
	ctx context.Context,
	fullapi v1api.FullNode,
	maddr address.Address,
) (string, error) {
	mi, err := fullapi.StateMinerInfo(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return "", fmt.Errorf("get miner info: %w", err)
	}

	secCounts, err := fullapi.StateMinerSectorCount(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return "", fmt.Errorf("get miner sector count: %w", err)
	}

	committed := types.BigMul(
		types.NewInt(secCounts.Live),
		types.NewInt(uint64(mi.SectorSize)),
	)

	return types.SizeStr(committed), nil
}

func computeUserDataSize(ctx context.Context, minerapi api.StorageMiner) (string, error) {
	list, err := minerapi.SectorsList(ctx)
	if err != nil {
		return "", fmt.Errorf("list sectors: %w", err)
	}

	var total uint64

	for _, sectorNum := range list {
		st, err := minerapi.SectorsStatus(ctx, abi.SectorNumber(sectorNum), true)
		if err != nil {
			return "", fmt.Errorf("get sector %d status: %w", sectorNum, err)
		}

		if st.State == api.SectorState(sealing.Removed) {
			continue
		}

		var sectorBytes uint64
		estimate := (st.Expiration-st.Activation <= 0) || sealing.IsUpgradeState(sealing.SectorState(st.State))

		if !estimate {
			rdw := big.Add(st.DealWeight, st.VerifiedDealWeight)
			sectorBytes = big.Div(rdw, big.NewInt(int64(st.Expiration-st.Activation))).Uint64()
		} else {
			for _, piece := range st.Pieces {
				if piece.DealInfo != nil {
					sectorBytes += uint64(piece.Piece.Size)
				}
			}
		}

		total += sectorBytes
	}

	return types.SizeStr(types.NewInt(total)), nil
}

func writeMinerInfoLog(path string, info *minerInfoSnapshot) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create miner info log dir: %w", err)
		}
	}

	lines := []string{
		fmt.Sprintf("ip=%s", info.IP),
		fmt.Sprintf("miner=%s", info.Miner),
		fmt.Sprintf("quality_adjusted_power=%s", info.QualityAdjustedPower),
		fmt.Sprintf("committed_space=%s", info.CommittedSpace),
		fmt.Sprintf("user_data_size=%s", info.UserDataSize),
	}

	data := strings.Join(lines, "\n") + "\n"

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("write temp miner info log: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace miner info log: %w", err)
	}

	return nil
}
