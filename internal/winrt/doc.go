package winrt

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -package winrt -class Windows.ApplicationModel.AppInstance -out app_instance.go
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -package winrt -class Windows.ApplicationModel.Activation.ShareTargetActivatedEventArgs -out share_args.go
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -package winrt -class Windows.ApplicationModel.DataTransfer.ShareTarget.ShareOperation -out share_op.go
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -package winrt -class Windows.ApplicationModel.DataTransfer.DataPackageView -out data_package.go
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -package winrt -class Windows.Storage.StorageFile -out storage_file.go
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -package winrt -class Windows.Storage.IStorageItem -out storage_item.go
