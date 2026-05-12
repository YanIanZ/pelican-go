package remote

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pelican-dev/panel/internal/models"
)

type ServerHandler struct {
	DB *gorm.DB
}

func (h *ServerHandler) List(c *gin.Context) {
	nodeID := c.GetUint("node_id")

	var servers []models.Server
	if err := h.DB.Where("node_id = ?", nodeID).
		Preload("Allocation").
		Preload("Egg").
		Preload("Owner").
		Find(&servers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": []gin.H{{"code": "ServerError", "status": "500", "detail": "Failed to list servers"}}})
		return
	}

	var eggIDs []uint
	var serverIDs []uint
	for _, s := range servers {
		eggIDs = append(eggIDs, s.EggID)
		serverIDs = append(serverIDs, s.ID)
	}

	var allEggVars []models.EggVariable
	if len(eggIDs) > 0 {
		h.DB.Where("egg_id IN ?", eggIDs).Find(&allEggVars)
	}

	var allServerVars []models.ServerVariable
	if len(serverIDs) > 0 {
		h.DB.Where("server_id IN ?", serverIDs).Find(&allServerVars)
	}

	var allAllocs []models.Allocation
	if len(serverIDs) > 0 {
		h.DB.Where("server_id IN ?", serverIDs).Find(&allAllocs)
	}

	var nodeIDs []uint
	for _, s := range servers {
		nodeIDs = append(nodeIDs, s.NodeID)
	}
	var allMountNodes []models.MountNode
	if len(nodeIDs) > 0 {
		h.DB.Where("node_id IN ?", nodeIDs).Find(&allMountNodes)
	}
	var allMountEggs []models.MountEgg
	if len(eggIDs) > 0 {
		h.DB.Where("egg_id IN ?", eggIDs).Find(&allMountEggs)
	}

	mountIDSet := make(map[uint]bool)
	for _, mn := range allMountNodes {
		for _, me := range allMountEggs {
			if mn.MountID == me.MountID {
				mountIDSet[mn.MountID] = true
				break
			}
		}
	}
	var allMounts []models.Mount
	for mid := range mountIDSet {
		var m models.Mount
		if h.DB.First(&m, mid).Error == nil {
			allMounts = append(allMounts, m)
		}
	}

	mountsByEggNode := make(map[uint]map[uint][]gin.H)
	for _, m := range allMounts {
		for _, me := range allMountEggs {
			if me.MountID == m.ID {
				for _, mn := range allMountNodes {
					if mn.MountID == m.ID {
						if mountsByEggNode[me.EggID] == nil {
							mountsByEggNode[me.EggID] = make(map[uint][]gin.H)
						}
						mountsByEggNode[me.EggID][mn.NodeID] = append(mountsByEggNode[me.EggID][mn.NodeID], gin.H{
							"source":    m.Source,
							"target":    m.Target,
							"read_only": m.ReadOnly,
						})
					}
				}
			}
		}
	}

	var data []gin.H
	for _, s := range servers {
		srv := s
		ev := filterEggVars(allEggVars, srv.EggID)
		sv := filterServerVars(allServerVars, srv.ID)
		al := filterAllocs(allAllocs, srv.ID)
		item := h.buildServerConfig(&srv, ev, sv, al)
		if mnts := mountsByEggNode[srv.EggID][srv.NodeID]; mnts != nil {
			if settings, ok := item["settings"].(gin.H); ok {
				settings["mounts"] = mnts
			}
		}
		data = append(data, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": gin.H{
			"pagination": gin.H{
				"total":        len(servers),
				"count":        len(servers),
				"per_page":     len(servers),
				"current_page": 0,
				"total_pages":  1,
			},
		},
	})
}

func filterEggVars(vars []models.EggVariable, eggID uint) []models.EggVariable {
	var out []models.EggVariable
	for _, v := range vars {
		if v.EggID == eggID {
			out = append(out, v)
		}
	}
	return out
}

func filterServerVars(vars []models.ServerVariable, serverID uint) []models.ServerVariable {
	var out []models.ServerVariable
	for _, v := range vars {
		if v.ServerID == serverID {
			out = append(out, v)
		}
	}
	return out
}

func filterAllocs(allocs []models.Allocation, serverID uint) []models.Allocation {
	var out []models.Allocation
	for _, a := range allocs {
		if a.ServerID != nil && *a.ServerID == serverID {
			out = append(out, a)
		}
	}
	return out
}

func (h *ServerHandler) Get(c *gin.Context) {
	uuid := c.Param("uuid")

	var server models.Server
	if err := h.DB.Where("uuid = ?", uuid).
		Preload("Allocation").
		Preload("Egg").
		Preload("Node").
		Preload("Owner").
		First(&server).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"errors": []gin.H{{"code": "HttpNotFoundException", "status": "404", "detail": "Server not found"}}})
		return
	}

	var eggVars []models.EggVariable
	h.DB.Where("egg_id = ?", server.EggID).Find(&eggVars)

	var serverVars []models.ServerVariable
	h.DB.Where("server_id = ?", server.ID).Find(&serverVars)

	var allocations []models.Allocation
	h.DB.Where("server_id = ?", server.ID).Find(&allocations)

	var mounts []gin.H
	var mountNodes []models.MountNode
	h.DB.Where("node_id = ?", server.NodeID).Find(&mountNodes)

	var mountEggs []models.MountEgg
	h.DB.Where("egg_id = ?", server.EggID).Find(&mountEggs)

	mountIDSet := make(map[uint]bool)
	for _, mn := range mountNodes {
		for _, me := range mountEggs {
			if mn.MountID == me.MountID {
				mountIDSet[mn.MountID] = true
				break
			}
		}
	}

	for mountID := range mountIDSet {
		var m models.Mount
		if h.DB.First(&m, mountID).Error == nil {
			mounts = append(mounts, gin.H{
				"source":    m.Source,
				"target":    m.Target,
				"read_only": m.ReadOnly,
			})
		}
	}

	config := h.buildServerConfig(&server, eggVars, serverVars, allocations)
	if settings, ok := config["settings"].(gin.H); ok {
		settings["mounts"] = mounts
	}

	c.JSON(http.StatusOK, config)
}

func miBToBytes(mib int64) int64 { return mib * 1048576 }

func (h *ServerHandler) buildServerConfig(server *models.Server, eggVars []models.EggVariable, serverVars []models.ServerVariable, allocs []models.Allocation) gin.H {
	envMap := make(map[string]string)
	for _, ev := range eggVars {
		val := ev.DefaultValue
		for _, sv := range serverVars {
			if sv.VariableID == ev.ID {
				val = sv.VariableValue
				break
			}
		}
		envMap[ev.EnvVariable] = val
	}

	invocation := server.Startup
	for k, v := range envMap {
		invocation = strings.ReplaceAll(invocation, "{{"+k+"}}", v)
	}

	mappings := make(map[string][]int)
	var defaultAlloc gin.H
	for _, a := range allocs {
		mappings[a.IP] = append(mappings[a.IP], a.Port)
		if server.AllocationID != nil && a.ID == *server.AllocationID {
			defaultAlloc = gin.H{"ip": a.IP, "port": a.Port}
		}
	}
	if defaultAlloc == nil && len(allocs) > 0 {
		defaultAlloc = gin.H{"ip": allocs[0].IP, "port": allocs[0].Port}
	}
	if len(mappings) == 0 {
		mappings[""] = []int{}
	}

	suspended := false
	if server.Status != nil && *server.Status == "suspended" {
		suspended = true
	}

	settings := gin.H{
		"id":        server.ID,
		"uuid":      server.UUID,
		"meta":      gin.H{"name": server.Name, "description": server.Description},
		"suspended": suspended,
		"environment": envMap,
		"invocation":  invocation,
		"skip_egg_scripts": server.SkipScripts,
		"build": gin.H{
			"memory_limit": miBToBytes(server.Memory),
			"swap":         server.Swap,
			"io_weight":    server.IO,
			"cpu_limit":    server.CPU,
			"threads":      server.Threads,
			"disk_space":   miBToBytes(server.Disk),
			"oom_killer":   server.OOMKiller,
		},
		"container": gin.H{
			"image":            server.Image,
			"requires_rebuild": false,
		},
		"allocations": gin.H{
			"force_outgoing_ip": server.Egg.ForceOutgoingIP,
			"default":           defaultAlloc,
			"mappings":          mappings,
		},
	}

	var startupDone []string
	if server.Egg.ConfigStartup != nil {
		var cfg struct {
			Done            []string `json:"done"`
			UserInteraction []string `json:"user_interaction"`
			StripAnsi       bool     `json:"strip_ansi"`
		}
		if err := json.Unmarshal([]byte(*server.Egg.ConfigStartup), &cfg); err == nil {
			startupDone = cfg.Done
		}
	}

	stopType := "command"
	stopValue := "stop"
	if server.Egg.ConfigStop != nil && *server.Egg.ConfigStop != "" {
		stopValue = *server.Egg.ConfigStop
	}

	var configs []gin.H
	if server.Egg.ConfigFiles != nil {
		var filesMap map[string]gin.H
		if err := json.Unmarshal([]byte(*server.Egg.ConfigFiles), &filesMap); err == nil {
			for file, cfg := range filesMap {
				configs = append(configs, gin.H{
					"file":   file,
					"parser": cfg["parser"],
					"find":   cfg["find"],
				})
			}
		}
	}

	return gin.H{
		"settings": settings,
		"process_configuration": gin.H{
			"startup": gin.H{
				"done":             startupDone,
				"user_interaction": []string{},
				"strip_ansi":      false,
			},
			"stop": gin.H{
				"type":  stopType,
				"value": stopValue,
			},
			"configs": configs,
		},
	}
}

func (h *ServerHandler) ResetState(c *gin.Context) {
	nodeID := c.GetUint("node_id")
	h.DB.Model(&models.Server{}).Where("node_id = ? AND status IN ?", nodeID, []string{"installing", "restoring_backup"}).Update("status", nil)
	c.JSON(http.StatusNoContent, nil)
}
