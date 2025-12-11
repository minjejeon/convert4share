export namespace main {

	export class Settings {
	    magickBinary: string;
	    ffmpegBinary: string;
	    maxSize: number;
	    hardwareAccelerator: string;
	    ffmpegCustomArgs: string;
	    defaultDestDir: string;
	    excludePatterns: string[];

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
	    }
	}

}

export namespace options {

	export class SecondInstanceData {
	    Args: string[];
	    WorkingDirectory: string;

	    static createFrom(source: any = {}) {
	        return new SecondInstanceData(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Args = source["Args"];
	        this.WorkingDirectory = source["WorkingDirectory"];
	    }
	}

}
