const chai = require('chai');
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

// Make sure you require the health check endpoint test so that it runs first
require('./health-check-test');


const {
  SERVICE_URL,
  REQUESTER_ID_HEADER,
} = require('./testcases.js');

const endpointsToTest = [
  {
    endpoint: '/user/foobar-models',
    methods: ['get', 'post'],
  },
  {
    endpoint: '/user/foobar-models/id',
    methods: ['patch', 'delete'],
  },
];
const improperIdHeaders = [
  'qwe',
  '',
  null,
];

describe("General Endpoint Tests", function() {

  describe("Endpoints With user-id Header Requirements Tests", function() {
    endpointsToTest.forEach(function({endpoint, methods}) {
      methods.forEach(function(method) {
        improperIdHeaders.forEach(function(header) {

          it(`should return error on endpoint '${method}' '${endpoint}' if header '${REQUESTER_ID_HEADER}' is improper: ${header}`, function(done) {
            chai.request(SERVICE_URL)[method](endpoint)
              .set(REQUESTER_ID_HEADER, header)
              .then(done, function(err) {
                chai.assert.equal(err.status, 400);
                done();
              })
              .catch(done);
          });

        });
      });
    });
  });

});
