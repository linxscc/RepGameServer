import React, { Fragment } from 'react'
import { StepProps } from '@/types'
import './steps2.css'

const Steps2: React.FC<StepProps> = (props) => {
  return (
    <div className="thq-section-padding">
      <div className="steps2-container2 thq-section-max-width">
        <div className="steps2-container3">
          <div className="steps2-container4">
            <h2 className="thq-heading-2">
              {props.step1Title ?? (
                <Fragment>
                  <span>Step 1</span>
                </Fragment>
              )}
            </h2>
            <span className="thq-body-small">
              {props.step1Description ?? (
                <Fragment>
                  <span>Step 1 description</span>
                </Fragment>
              )}
            </span>
          </div>
          <div className="steps2-container5">
            <h2 className="thq-heading-2">
              {props.step2Title ?? (
                <Fragment>
                  <span>Step 2</span>
                </Fragment>
              )}
            </h2>
            <span className="thq-body-small">
              {props.step2Description ?? (
                <Fragment>
                  <span>Step 2 description</span>
                </Fragment>
              )}
            </span>
          </div>
          <div className="steps2-container6">
            <h2 className="thq-heading-2">
              {props.step3Title ?? (
                <Fragment>
                  <span>Step 3</span>
                </Fragment>
              )}
            </h2>
            <span className="thq-body-small">
              {props.step3Description ?? (
                <Fragment>
                  <span>Step 3 description</span>
                </Fragment>
              )}
            </span>
          </div>
          <div className="steps2-container7">
            <h2 className="thq-heading-2">
              {props.step4Title ?? (
                <Fragment>
                  <span>Step 4</span>
                </Fragment>
              )}
            </h2>
            <span className="thq-body-small">
              {props.step4Description ?? (
                <Fragment>
                  <span>Step 4 description</span>
                </Fragment>
              )}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Steps2
