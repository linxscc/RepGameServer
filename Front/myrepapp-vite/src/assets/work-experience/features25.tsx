import React, { Fragment } from 'react'
import { FeatureProps } from '@/types'
import './features25.css'

const Features25: React.FC<FeatureProps> = (props) => {
  return (
    <div className="thq-section-padding">
      <div className="features25-container2 thq-section-max-width">
        <div className="features25-tabs-menu">
          <div className="features25-tab-horizontal1">
            <div className="features25-divider-container1">
              <div className="features25-container3"></div>
            </div>
            <div className="features25-content1">
              <h2 className="thq-heading-2">
                {props.feature1Title ?? (
                  <Fragment>
                    <span>Feature 1</span>
                  </Fragment>
                )}
              </h2>
              <span className="thq-body-small">
                {props.feature1Description ?? (
                  <Fragment>
                    <span>Feature 1 description</span>
                  </Fragment>
                )}
              </span>
            </div>
          </div>
          <div className="features25-tab-horizontal2">
            <div className="features25-divider-container2">
              <div className="features25-container4"></div>
            </div>
            <div className="features25-content2">
              <h2 className="thq-heading-2">
                {props.feature2Title ?? (
                  <Fragment>
                    <span>Feature 2</span>
                  </Fragment>
                )}
              </h2>
              <span className="thq-body-small">
                {props.feature2Description ?? (
                  <Fragment>
                    <span>Feature 2 description</span>
                  </Fragment>
                )}
              </span>
            </div>
          </div>
          <div className="features25-tab-horizontal3">
            <div className="features25-divider-container3">
              <div className="features25-container5"></div>
            </div>
            <div className="features25-content3">
              <h2 className="thq-heading-2">
                {props.feature3Title ?? (
                  <Fragment>
                    <span>Feature 3</span>
                  </Fragment>
                )}
              </h2>
              <span className="thq-body-small">
                {props.feature3Description ?? (
                  <Fragment>
                    <span>Feature 3 description</span>
                  </Fragment>
                )}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Features25
