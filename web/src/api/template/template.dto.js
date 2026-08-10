export class TemplateDto {
    constructor(data) {
        this.code = data.code || null;
        this.name = data.name;
        this.description = data.description;
        this.params = data.params || {};
        this.subject = data.subject;
        this.body = data.body;
    }
}

export class TemplateFinderResponseDto {
    constructor(backendData) {
        const rawData = backendData || {};

        this.messages = Array.isArray(rawData.templates)
            ? rawData.messages.map(m => new TemplateDto(m))
            : [];

        const pag = rawData.pagination || {};
        this.pagination = {
            currentPage: Number(pag.current_page ?? 1),
            perPage: Number(pag.per_page ?? 10),
            total: Number(pag.total ?? 0),
            totalPages: Number(pag.total_pages ?? 1)
        };
    }
}