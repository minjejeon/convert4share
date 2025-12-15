export namespace main {

	export class JobStatus {
	    id: string;
	    file: string;
	    status: string;
	    progress: number;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new JobStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.file = source["file"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.error = source["error"];
	    }
	}
	export class Settings {
	    magickBinary: string;
	    ffmpegBinary: string;
	    maxSize: number;
	    hardwareAccelerator: string;
	    ffmpegCustomArgs: string;
	    defaultDestDir: string;
	    excludePatterns: string[];
	    videoQuality: string;

	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.magickBinary = source["magickBinary"];
	        this.ffmpegBinary = source["ffmpegBinary"];
	        this.maxSize = source["maxSize"];
	        this.hardwareAccelerator = source["hardwareAccelerator"];
	        this.ffmpegCustomArgs = source["ffmpegCustomArgs"];
	        this.defaultDestDir = source["defaultDestDir"];
	        this.excludePatterns = source["excludePatterns"];
	        this.videoQuality = source["videoQuality"];
	    }
	}

}
