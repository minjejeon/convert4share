//go:build windows && winrt

package share

import (
	"log"
	"os"
	"time"

	"github.com/minjejeon/convert4share/internal/winrt"
	"github.com/saltosystems/winrt-go/windows/foundation"
)

// CheckActivation checks if the app was activated via Share Target.
// It returns true if handled (files added to args), false otherwise.
func CheckActivation() bool {
	// 1. Get Activation Args
	args, err := winrt.AppInstance_GetActivatedEventArgs()
	if err != nil {
		return false
	}

	kind, err := args.GetKind()
	if err != nil {
		return false
	}

	// ActivationKind_ShareTarget == 5
	if kind != 5 {
		return false
	}

	log.Println("App activated via Share Target")

	// 2. Cast to ShareTargetActivatedEventArgs
	// Assuming winrt-go generated the helper method FromIActivatedEventArgs
	shareArgs := winrt.ShareTargetActivatedEventArgs_FromIActivatedEventArgs(args)
	if shareArgs == nil {
		log.Println("Failed to cast to ShareTargetActivatedEventArgs")
		return false
	}

	// 3. Get Share Operation
	op, err := shareArgs.GetShareOperation()
	if err != nil {
		log.Println("Failed to get ShareOperation:", err)
		return false
	}

	// 4. Get DataPackageView
	data, err := op.GetData()
	if err != nil {
		log.Println("Failed to get Data:", err)
		return false
	}

	// 5. Get Storage Items (Async)
	// We assume GetStorageItemsAsync is generated and returns an IAsyncOperation
	asyncOp, err := data.GetStorageItemsAsync()
	if err != nil {
		log.Println("Failed to call GetStorageItemsAsync:", err)
		return false
	}

	// 6. Poll for completion
	log.Println("Waiting for storage items...")
	for {
		status, err := asyncOp.GetStatus()
		if err != nil {
			log.Println("Failed to get async status:", err)
			return false
		}

		if status == foundation.AsyncStatus_Completed {
			break
		}
		if status == foundation.AsyncStatus_Error || status == foundation.AsyncStatus_Canceled {
			log.Println("Async operation failed or canceled")
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 7. Get Results
	items, err := asyncOp.GetResults()
	if err != nil {
		log.Println("Failed to get results:", err)
		return false
	}

	size, err := items.GetSize()
	if err != nil {
		log.Println("Failed to get size:", err)
		return false
	}

	log.Printf("Received %d items", size)

	var newFiles []string
	for i := uint32(0); i < size; i++ {
		item, err := items.GetAt(i)
		if err != nil {
			continue
		}

		// IStorageItem has Path property
		path, err := item.GetPath()
		if err != nil {
			log.Println("Failed to get path for item:", i)
			continue
		}

		newFiles = append(newFiles, path)
	}

	if len(newFiles) > 0 {
		log.Printf("Adding shared files to args: %v", newFiles)
		os.Args = append(os.Args, newFiles...)
	}

	// 8. Report Completed
	op.ReportCompleted()

	return true
}
