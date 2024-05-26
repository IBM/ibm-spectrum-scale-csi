package scale

import (
	"context"
	"sync"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/utils"
	"k8s.io/klog/v2"
)

const (
	createVolume      = "CreateVolume"
	createSnapshot    = "CreateSnapshot"
	deleteVolume      = "DeleteVolume"
	deleteSnapshot    = "DeleteSnapshot"
	retryInterval     = 10
	retryCount        = 2
	createOpInitCount = 1
)

var createVolumeRefLock map[string]int
var createSnapshotRefLock map[string]int
var cgSnapLock map[string]string
var cgSnapMutex sync.Mutex


func CgSnapshotLock(ctx context.Context, targetPath string, snapExists bool) bool {
	cgSnapMutex.Lock()
	defer cgSnapMutex.Unlock()
	return cgParallelSnapshotLock(ctx,targetPath, snapExists )	
}

func cgParallelSnapshotLock(ctx context.Context, targetPath string, snapExists bool) bool {

        if len(createVolumeRefLock) == 0 {
                createVolumeRefLock = make(map[string]int)
        }

        if len(createSnapshotRefLock) == 0 {
                createSnapshotRefLock = make(map[string]int)
        }

        if len(cgSnapLock) == 0 {
                cgSnapLock = make(map[string]string)
        }

        lockingModule := utils.GetModuleName(ctx)

        createVolCount, _ := createVolumeRefLock[targetPath]
        createSnapCount, _ := createSnapshotRefLock[targetPath]
        moduleName, exists := cgSnapLock[targetPath]
        if createVolCount > 0 || createSnapCount > 0 {
                if lockingModule == createVolume {
                        createVolumeRefLock[targetPath]++
                        cgSnapLock[targetPath] = lockingModule
                        return true
                } else if lockingModule == createSnapshot {
                        if !snapExists {
                                klog.Infof("[%s] Snap doesn't exist and lock already acquired by another snapshot request", utils.GetLoggerId(ctx))
                                return false
                        } else {
                                createSnapshotRefLock[targetPath]++
                                cgSnapLock[targetPath] = lockingModule
                                return true
                        }
                } else {
                        klog.Infof("[%s] Delete operation is trying to acquire lock while create action is in progress", utils.GetLoggerId(ctx))
			return false
                }
        } else {
                if exists && moduleName != "" {
                        if (moduleName == deleteVolume || moduleName == deleteSnapshot) && (lockingModule == createVolume || lockingModule == createSnapshot) {
                                klog.Infof("[%s] Delete operation acquired the lock, create operation retrying", utils.GetLoggerId(ctx))
				return false
                        }
                } else {
                        if lockingModule == createVolume {
                                createVolumeRefLock[targetPath] = createOpInitCount
                        } else if lockingModule == createSnapshot {
                                createSnapshotRefLock[targetPath] = createOpInitCount
                        } else {
                                klog.Infof("[%s] Delete operation acquired the lock", utils.GetLoggerId(ctx))
                        }
                }
                cgSnapLock[targetPath] = lockingModule
        }

        klog.V(4).Infof("[%s] The target path is locked for %s: [%s]", utils.GetLoggerId(ctx), utils.GetModuleName(ctx), targetPath)
        return true
}


func CgSnapshotUnlock(ctx context.Context, targetPath string) {
	cgSnapMutex.Lock()
	defer cgSnapMutex.Unlock()
	cgParallelSnapshotUnlock(ctx, targetPath)	
}

func cgParallelSnapshotUnlock(ctx context.Context, targetPath string) {
        moduleName := utils.GetModuleName(ctx)
        if moduleName == createVolume {
                if createVolumeRefLock[targetPath] > 0 {
                        createVolumeRefLock[targetPath]--
                } else {
                        delete(createVolumeRefLock, targetPath)
                }
        } else if moduleName == createSnapshot {
                if createSnapshotRefLock[targetPath] > 0 {
                        createSnapshotRefLock[targetPath]--
                } else {
                        delete(createSnapshotRefLock, targetPath)
                }
        } else {
                klog.Infof("[%s] Delete operation released the lock", utils.GetLoggerId(ctx))
        }
        delete(cgSnapLock, targetPath)
        klog.V(4).Infof("[%s] The target path is unlocked for %s: [%s]", utils.GetLoggerId(ctx), utils.GetModuleName(ctx), targetPath)
}

/*func retrySnapLock(ctx context.Context, targetPath, lockingModule string, snapExists bool) error {
        for i := 0; i < retryCount; i++ {
                time.Sleep(retryInterval * time.Second)
                if cgParallelSnapshotLock(ctx, targetPath, snapExists) {
                        klog.Infof("[%s] retry attempt for %s operation", utils.GetLoggerId(ctx), utils.GetModuleName(ctx))
                        return nil
                } else {
                        klog.Errorf("[%s] Failed to lock the target path after retry attampts", utils.GetLoggerId(ctx))
                        return status.Error(codes.Internal, fmt.Sprintf("Failed to lock the target path [%s] by %s after retry attampts", targetPath, utils.GetModuleName(ctx)))
                }
        }
        return nil
}*/
