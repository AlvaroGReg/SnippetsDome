export namespace domain {
	
	export class CreateSnippetInput {
	    title: string;
	    language: string;
	    code: string;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateSnippetInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.language = source["language"];
	        this.code = source["code"];
	        this.tags = source["tags"];
	    }
	}
	export class Snippet {
	    id: string;
	    title: string;
	    language: string;
	    code: string;
	    tags: string[];
	    createdAt: string;
	    favorite: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Snippet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.language = source["language"];
	        this.code = source["code"];
	        this.tags = source["tags"];
	        this.createdAt = source["createdAt"];
	        this.favorite = source["favorite"];
	    }
	}

}

