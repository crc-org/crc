<xsl:stylesheet version="1.0"
xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
<xsl:output method="xml" version="1.0" encoding="UTF-8" indent="yes"/>
<xsl:strip-space elements="*"/>

<!-- identity transform -->
<xsl:template match="@*|node()">
    <xsl:copy>
            <xsl:apply-templates select="@*|node()"/>
    </xsl:copy>
</xsl:template>

<xsl:template match="failure">
    <xsl:copy>
        <xsl:value-of select="@message"/>
    </xsl:copy>
</xsl:template>

</xsl:stylesheet>
