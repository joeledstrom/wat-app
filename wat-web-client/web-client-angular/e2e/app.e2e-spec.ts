import { WebClientAngularPage } from './app.po';

describe('web-client-angular App', () => {
  let page: WebClientAngularPage;

  beforeEach(() => {
    page = new WebClientAngularPage();
  });

  it('should display message saying app works', () => {
    page.navigateTo();
    expect(page.getParagraphText()).toEqual('app works!');
  });
});
